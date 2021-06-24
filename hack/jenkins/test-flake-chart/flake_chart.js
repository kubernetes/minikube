
// Displays an error message to the UI. Any previous message will be erased.
function displayError(message) {
  console.error(message);
}

// Creates a generator that reads the response body one line at a time.
async function* bodyByLinesIterator(response) {
  // TODO: Replace this with something that actually reads the body line by line
  // (since the file can be big).
  const lines = (await response.text()).split("\n");
  for (let line of lines) {
    // Skip any empty lines (most likely at the end).
    if (line !== "") {
      yield line;
    }
  }
}

// Determines whether `str` matches at least one value in `enumObject`.
function isValidEnumValue(enumObject, str) {
  for (const enumKey in enumObject) {
    if (enumObject[enumKey] === str) {
      return true;
    }
  }
  return false;
}

// Enum for test status.
const testStatus = {
  PASSED: "Passed",
  FAILED: "Failed",
  SKIPPED: "Skipped"
}

async function loadTestData() {
  const response = await fetch("data.csv");
  if (!response.ok) {
    const responseText = await response.text();
    throw `Failed to fetch data from GCS bucket. Error: ${responseText}`;
  }

  const lines = bodyByLinesIterator(response);
  // Consume the header to ensure the data has the right number of fields.
  const header = (await lines.next()).value;
  if (header.split(",").length != 6) {
    throw `Fetched CSV data contains wrong number of fields. Expected: 6. Actual Header: "${header}"`;
  }

  const testData = [];
  let lineData = ["", "", "", "", "", ""];
  for await (const line of lines) {
    let splitLine = line.split(",");
    if (splitLine.length != 6) {
      console.warn(`Found line with wrong number of fields. Actual: ${splitLine.length} Expected: 6. Line: "${line}"`);
      continue;
    }
    splitLine = splitLine.map((value, index) => value === "" ? lineData[index] : value);
    lineData = splitLine;
    if (!isValidEnumValue(testStatus, splitLine[4])) {
      console.warn(`Invalid test status provided. Actual: ${splitLine[4]} Expected: One of ${Object.values(testStatus).join(", ")}`);
      continue;
    }
    testData.push({
      commit: splitLine[0],
      date: new Date(splitLine[1]),
      environment: splitLine[2],
      name: splitLine[3],
      status: splitLine[4],
      duration: Number(splitLine[5]),
    });
  }
  if (testData.length == 0) {
    throw "Fetched CSV data is empty or poorly formatted.";
  }
  return testData;
}

Array.prototype.sum = function() {
  return this.reduce((sum, value) => sum + value, 0);
};

// Computes the average of an array of numbers.
Array.prototype.average = function () {
  return this.length === 0 ? 0 : (this.sum() / this.length);
};

// Groups array elements by keys obtained through `keyGetter`.
Array.prototype.groupBy = function (keyGetter) {
  return Array.from(this.reduce((mapCollection, element) => {
    const key = keyGetter(element);
    if (mapCollection.has(key)) {
      mapCollection.get(key).push(element);
    } else {
      mapCollection.set(key, [element]);
    }
    return mapCollection;
  }, new Map()).values());
};

// Parse URL search `query` into [{key, value}].
function parseUrlQuery(query) {
  if (query[0] === '?') {
    query = query.substring(1);
  }
  return Object.fromEntries((query === "" ? [] : query.split("&")).map(element => {
    const keyValue = element.split("=");
    return [unescape(keyValue[0]), unescape(keyValue[1])];
  }));
}

// Takes a set of test runs (all of the same test), and aggregates them into one element per date.
function aggregateRuns(testRuns) {
  return testRuns
    // Group runs by the date it ran.
    .groupBy(run => run.date.getTime())
    // Sort by run date, past to future.
    .sort((a, b) => a[0].date - b[0].date)
    // Map each group to all variables need to format the rows.
    .map(tests => ({
      date: tests[0].date, // Get one of the dates from the tests (which will all be the same).
      flakeRate: tests.map(test => test.status === testStatus.FAILED ? 100 : 0).average(), // Compute average of runs where FAILED counts as 100%.
      duration: tests.map(test => test.duration).average(), // Compute average duration of runs.
      commitHashes: tests.map(test => ({ // Take all hashes, statuses, and durations of tests in this group.
        hash: test.commit,
        status: test.status,
        duration: test.duration
      })).groupBy(run => run.hash).map(runsWithSameHash => ({
        hash: runsWithSameHash[0].hash,
        failures: runsWithSameHash.map(run => run.status === testStatus.FAILED ? 1 : 0).sum(),
        runs: runsWithSameHash.length,
        duration: runsWithSameHash.map(run => run.duration).average(),
      }))
    }));
}

const hashToLink = (hash, environment) => `https://storage.googleapis.com/minikube-builds/logs/master/${hash.substring(0,7)}/${environment}.html`;

function displayTestAndEnvironmentChart(testData, testName, environmentName) {
  const data = new google.visualization.DataTable();
  data.addColumn('date', 'Date');
  data.addColumn('number', 'Flake Percentage');
  data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
  data.addColumn('number', 'Duration');
  data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });

  const testRuns = testData
    // Filter to only contain unskipped runs of the requested test and requested environment.
    .filter(test => test.name === testName && test.environment === environmentName && test.status !== testStatus.SKIPPED);

  data.addRows(
    aggregateRuns(testRuns)
      .map(groupData => [
        groupData.date,
        groupData.flakeRate,
        `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>${groupData.date.toString()}</b><br>
          <b>Flake Percentage:</b> ${groupData.flakeRate.toFixed(2)}%<br>
          <b>Hashes:</b><br>
          ${groupData.commitHashes.map(({ hash, failures, runs }) => `  - <a href="${hashToLink(hash, environmentName)}">${hash}</a> (Failures: ${failures}/${runs})`).join("<br>")}
        </div>`,
        groupData.duration,
        `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>${groupData.date.toString()}</b><br>
          <b>Average Duration:</b> ${groupData.duration.toFixed(2)}s<br>
          <b>Hashes:</b><br>
          ${groupData.commitHashes.map(({ hash, runs, duration }) => `  - <a href="${hashToLink(hash, environmentName)}">${hash}</a> (Average of ${runs}: ${duration.toFixed(2)}s)`).join("<br>")}
        </div>`,
      ])
  );

  const options = {
    title: `Flake rate and duration by day of ${testName} on ${environmentName}`,
    width: window.innerWidth,
    height: window.innerHeight,
    pointSize: 10,
    pointShape: "circle",
    series: {
      0: { targetAxisIndex: 0 },
      1: { targetAxisIndex: 1 },
    },
    vAxes: {
      0: { title: "Flake rate", minValue: 0, maxValue: 100 },
      1: { title: "Duration (seconds)" },
    },
    colors: ['#dc3912', '#3366cc'],
    tooltip: { trigger: "selection", isHtml: true }
  };
  const chart = new google.visualization.LineChart(document.getElementById('chart_div'));
  chart.draw(data, options);
}

function displayEnvironmentChart(testData, environmentName) {
  // Number of days to use to look for "flaky-est" tests.
  const dateRange = 15;
  // Number of tests to display in chart.
  const topFlakes = 10;

  const testRuns = testData
    // Filter to only contain unskipped runs of the requested test and requested environment.
    .filter(test => test.environment === environmentName && test.status !== testStatus.SKIPPED)
    .groupBy(test => test.name);

  const aggregatedRuns = new Map(testRuns.map(test => [
    test[0].name,
    new Map(aggregateRuns(test)
      .map(runDate => [ runDate.date.getTime(), runDate ]))]));
  const uniqueDates = new Set();
  for (const [_, runDateMap] of aggregatedRuns) {
    for (const [dateTime, _] of runDateMap) {
      uniqueDates.add(dateTime);
    }
  }
  const orderedDates = Array.from(uniqueDates).sort();
  const recentDates = orderedDates.slice(-dateRange);
  
  const recentFlakePercentage = Array.from(aggregatedRuns).map(([testName, data]) => {
    const {flakeCount, totalCount} = recentDates.map(date => {
      const dateInfo = data.get(date);
      return dateInfo === undefined ? null : {
        flakeRate: dateInfo.flakeRate,
        runs: dateInfo.commitHashes.length
      };
    }).filter(dateInfo => dateInfo != null)
      .reduce(({flakeCount, totalCount}, {flakeRate, runs}) => ({
        flakeCount: flakeRate * runs + flakeCount,
        totalCount: runs + totalCount
      }), {flakeCount: 0, totalCount: 0});
    return {
      testName,
      flakeRate: totalCount === 0 ? 0 : flakeCount / totalCount,
    };
  });

  const recentTopFlakes = recentFlakePercentage
    .sort((a, b) => b.flakeRate - a.flakeRate)
    .slice(0, topFlakes)
    .map(({testName}) => testName);

  const data = new google.visualization.DataTable();
  data.addColumn('date', 'Date');
  for (const name of recentTopFlakes) {
    data.addColumn('number', `Flake Percentage - ${name}`);
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
  }
  data.addRows(
    orderedDates.map(dateTime => [new Date(dateTime)].concat(recentTopFlakes.map(name => {
      const data = aggregatedRuns.get(name).get(dateTime);
      return data !== undefined ? [
        data.flakeRate,
        `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>${data.date.toString()}</b><br>
          <b>Flake Percentage:</b> ${data.flakeRate.toFixed(2)}%<br>
          <b>Hashes:</b><br>
          ${data.commitHashes.map(({ hash, failures, runs }) => `  - <a href="${hashToLink(hash, environmentName)}">${hash}</a> (Failures: ${failures}/${runs})`).join("<br>")}
        </div>`
      ] : [null, null];
    })).flat())
  );
  const options = {
    title: `Flake rate by day of top ${topFlakes} of recent test flakiness (past ${dateRange} days) on ${environmentName}`,
    width: window.innerWidth,
    height: window.innerHeight,
    pointSize: 10,
    pointShape: "circle",
    vAxes: {
      0: { title: "Flake rate", minValue: 0, maxValue: 100 },
    },
    tooltip: { trigger: "selection", isHtml: true }
  };
  const chart = new google.visualization.LineChart(document.getElementById('chart_div'));
  chart.draw(data, options);
}

async function init() {
  google.charts.load('current', { 'packages': ['corechart'] });
  let testData;
  try {
    // Wait for Google Charts to load, and for test data to load.
    // Only store the test data (at index 1) into `testData`.
    testData = (await Promise.all([
      new Promise(resolve => google.charts.setOnLoadCallback(resolve)),
      loadTestData()
    ]))[1];
  } catch (err) {
    displayError(err);
    return;
  }

  const query = parseUrlQuery(window.location.search);
  const desiredTest = query.test, desiredEnvironment = query.env || "";

  if (desiredTest === undefined) {
    displayEnvironmentChart(testData, desiredEnvironment);
  } else {
    displayTestAndEnvironmentChart(testData, desiredTest, desiredEnvironment);
  }
}

init();
