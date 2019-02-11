#!/usr/bin/env python
"""This script will execute a set of named minikube tests,
   gather the results, logs, and artifacts into a named GCS
   bucket for presentation in k8s testgrid:
   https://k8s-testgrid.appspot.com
"""

import os, sys, json, re, argparse, calendar, time, subprocess, shlex

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
  """
  started_json_file = open('%s/started.json' % outdir, 'w')
  finished_json_file = open('%s/finished.json' % outdir, 'w')
  junit_xml_file = open('%s/artifacts/junit_runner.xml' % outdir, 'w')

  failures = 0
  testxml = ""
  for test in test_results:
    testxml += '<testcase classname="%s" name="%s" time="%s">' % (test['classname'], test['name'], test['time'])
    if test['status'] == 'FAIL':
        failures += 1
        testxml += '<failure message="Test Failed" />'
    testxml += '</testcase>\n'
  junit_xml_file.write('<testsuite failures="%s" tests="%s">\n' % (failures, len(test_results)))
  junit_xml_file.write(testxml)
  junit_xml_file.write('</testsuite>')
  junit_xml_file.close()

  started_json_file.write(json.dumps(started)) 
  started_json_file.close()
  finished_json_file.write(json.dumps(finished)) 
  finished_json_file.close()

  return

def upload_results(outdir, test, buildnum, bucket):
  """ push the contents of gcs_out/* into bucket/test/logs/buildnum"""
  classname = os.path.basename(test).split('.')[0]
  args = shlex.split("gsutil cp -R gcs_out/ gs://%s/logs/%s/%s" % (bucket, classname, buildnum))
  p = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
  for line in p.stdout:
    print line

def gather_artifacts():
  """ gather any named or default artifacts into artifacts/ """
  return

def run_tests(test, build_log, exit_status, started, finished, test_results):
  """ execute the test, grab the start time, finish time, build logs and exit status
      Pull test results and important information out of the build log 
      test results format should be:
      === RUN   TestFunctional/Mounting
      --- PASS: TestFunctional (42.87s)
          --- PASS: TestFunctional/Status (2.33s)
          --- FAIL: SOMETESTSUITE/TESTNAME (seconds)
  """
  classname = os.path.basename(test).split('.')[0]
  build_log_file = open(build_log, 'w')
  args = shlex.split('bash -x %s' % test)
  p = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
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
  parser.add_argument('--test', required=True, help='give full path to test script you want to run')
  parser.add_argument('--buildnum', required=True, help='buildnumber for uploading to GCS')
  parser.add_argument('--bucket', default="k8s-minikube-prow", help='Name of the GCS bucket to upload to.  Default: k8s-minkube-prow')
  parser.add_argument('--outdir', default="gcs_out", help='Path of the directory to store all results, artifacts, and logs')
  parser.add_argument('--artifact', help='SRCPATH:TARGETPATH use to specify a file that needs to be uploaded into GCS bucket')
  args = parser.parse_args()

  if not os.path.exists(args.outdir):
    os.makedirs(args.outdir)
    os.makedirs('%s/artifacts' % args.outdir)

  build_log="%s/build_log.txt" % (args.outdir)
  exit_status=""
  started={"timestamp":calendar.timegm(time.gmtime())}
  finished={}
  test_results=[]

  run_tests(args.test, build_log, exit_status, started, finished, test_results)

  finished['timestamp'] = calendar.timegm(time.gmtime())
  if exit_status != "0":
    finished['passed'] = "true"
    finished['result'] = "SUCCESS"
  else:
    finished['passed'] = "false"
    finished['result'] = "FAIL"

  write_results(args.outdir, started, finished, test_results)
  upload_results(args.outdir, args.test, args.buildnum, args.bucket)

if __name__ == '__main__':
  main(sys.argv)
