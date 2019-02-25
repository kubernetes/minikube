#!/usr/bin/env python

# Copyright 2019 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""This script will execute a set of named minikube tests,
   gather the results, logs, and artifacts into a named GCS
   bucket for presentation in k8s testgrid:
   https://k8s-testgrid.appspot.com
"""

import os, sys, json, re, argparse, calendar, time, subprocess, shlex

def get_classname(test_script):
  """ parse out the test classname from the full path of the test script"""
  classname = os.path.basename(test).split('.')[0]
  return classname

def write_results(outdir, started, finished, test_results):
  """ write current results into artifacts/junit_runner.xml 
      format:
      <testsuite failures="XXX" tests="YYY" time="ZZZ">
         <testcase classname="SUITENAME" name="TESTNAME" time="ZZZ1" />
         <testcase classname="SUITENAME" name="TESTNAME" time="ZZZ1" />
         ...
      </testsuite>

      write the started.json and finish.json files 
      format:
      started.json: {"timestamp":STARTTIMEINSECONDSINCEEPOCH}
      finished.json: {"timestamp":FINISHTIMEINSECONDSINCEEPOCH,
                      "passed":FINALRESULT,
                      "result":SUCCESS|FAIL,
                      "metadata":{}
                      }
     Args:
       outdir: a string containing the results storage directory
       started: a dict containing the starting data
       finished: a dict containing the finished data
       tests_results: a list of dicts containing test results
  """
  started_json = open(os.path.join(outdir, "started.json"), 'w')
  finished_json = open(os.path.join(outdir, "finished.json"), 'w')
  junit_xml = open(os.path.join(outdir, "artifacts", "junit_runner.xml"), 'w')

  failures = 0
  testxml = ""
  for test in test_results:
    testxml += '<testcase classname="%s" name="%s" time="%s">' % (test['classname'], test['name'], test['time'])
    if test['status'] == 'FAIL':
        failures += 1
        testxml += '<failure message="Test Failed" />'
    testxml += '</testcase>\n'
  junit_xml.write('<testsuite failures="%s" tests="%s">\n' % (failures, len(test_results)))
  junit_xml.write(testxml)
  junit_xml.write('</testsuite>')
  junit_xml.close()

  started_json.write(json.dumps(started)) 
  started_json.close()
  finished_json.write(json.dumps(finished)) 
  finished_json.close()

  return

def upload_results(outdir, test_script, buildnum, bucket):
  """ push the contents of gcs_out/* into bucket/test/logs/buildnum

      Args:
       outdir: a string containing the results storage directory
       test_script: a string containing path to the test script
       buildnum: a string containing the buildnum
       bucket: a string containing the bucket to upload results to
  """
  classname = get_classname(test_script)
  args = shlex.split("gsutil cp -R gcs_out/ gs://%s/logs/%s/%s" % (bucket, classname, buildnum))
  p = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
  for line in p.stdout:
    print line

def run_tests(test_script, log_path, exit_status, started, finished, test_results):
  """ execute the test script, grab the start time, finish time, build logs and exit status
      Pull test results and important information out of the build log 
      test results format should be:
      === RUN   TestFunctional/Mounting
      --- PASS: TestFunctional (42.87s)
          --- PASS: TestFunctional/Status (2.33s)
          --- FAIL: SOMETESTSUITE/TESTNAME (seconds)

      Args:
       test_script: a string containing path to the test script
       build_log: a string containing path to the build_log
       exit_status: a string that will contain the test script's exit_status
       started: a dict containing the starting data
       finished: a dict containing the finished data
       tests_results: a list of dicts containing test results
  """
  classname = get_classname(test_script)
  build_log_file = open(log_path, 'w')
  p = subprocess.Popen(['bash','-x',test_script], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
  for line in p.stdout:
    build_log_file.write(line)
    print line.rstrip()
    if '--- PASS' in line:
      match = re.match('.*--- PASS: ([^ ]+) \(([0-9.]+)s\)', line)
      (name, seconds) = match.group(1, 2)
      test_results.append({"name":name, "classname":classname, "time":seconds, "status":"PASS"})
    if '--- FAIL' in line:
      match = re.match('.*--- FAIL: ([^ ]+) \(([0-9.]+)s\)', line)
      (name, seconds) = match.group(1, 2)
      test_results.append({"name":name, "classname":classname, "time":seconds, "status":"FAIL"})
  build_log_file.close()
  return

def main(argv):

  parser = argparse.ArgumentParser(description='Run tests and upload results to GCS bucket', usage='./run_tests.py --test path/to/test.sh')
  parser.add_argument('--test', required=True, help='full path to test script you want to run')
  parser.add_argument('--build-num', dest="buildnum", required=True, help='buildnumber for uploading to GCS')
  parser.add_argument('--bucket', default="k8s-minikube-prow", help='Name of the GCS bucket to upload to.  Default: k8s-minkube-prow')
  parser.add_argument('--out-dir', dest="outdir", default="gcs_out", help='Path of the directory to store all results, artifacts, and logs')
  args = parser.parse_args()

  if not os.path.exists(args.outdir):
    os.makedirs(os.path.join(args.outdir, "artifacts"))

  build_log = os.path.join(args.outdir, "build_log.txt")
  exit_status = ""
  started = {"timestamp":calendar.timegm(time.gmtime())}
  finished = {}
  test_results = []

  run_tests(args.test, build_log, exit_status, started, finished, test_results)

  finished['timestamp'] = calendar.timegm(time.gmtime())
  #if the test script in run_tests exits with a non-zero status then mark the test run as FAILED
  if exit_status != "0":
    finished['passed'] = "false"
    finished['result'] = "FAIL"
  else:
    finished['passed'] = "true"
    finished['result'] = "SUCCESS"

  write_results(args.outdir, started, finished, test_results)
  upload_results(args.outdir, args.test, args.buildnum, args.bucket)

if __name__ == '__main__':
  main(sys.argv)
