
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
  for await (const line of lines) {
    const splitLine = line.split(",");
    if (splitLine.length != 6) {
      console.warn(`Found line with wrong number of fields. Actual: ${splitLine.length} Expected: 6. Line: "${line}"`);
      continue;
    }
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

// Computes the average of an array of numbers.
Array.prototype.average = function () {
  return this.length === 0 ? 0 : this.reduce((sum, value) => sum + value, 0) / this.length;
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

  const data = new google.visualization.DataTable();
  data.addColumn('date', 'Date');
  data.addColumn('number', 'Flake Percentage');
  data.addColumn({ type: 'string', label: 'Commit Hash', role: 'tooltip', 'p': { 'html': true } });

  const desiredTest = "TestFunctional/parallel/LogsCmd", desiredEnvironment = "Docker_Linux_containerd";

  const groups = testData
    // Filter to only contain unskipped runs of the requested test and requested environment.
    .filter(test => test.name === desiredTest && test.environment === desiredEnvironment && test.status !== testStatus.SKIPPED)
    .groupBy(test => test.date.getTime());

  data.addRows(
    groups
      // Sort by run date, past to future.
      .sort((a, b) => a[0].date - b[0].date)
      // Map each group to all variables need to format the rows.
      .map(tests => ({
        date: tests[0].date, // Get one of the dates from the tests (which will all be the same).
        flakeRate: tests.map(test => test.status === testStatus.FAILED ? 100 : 0).average(), // Compute average of runs where FAILED counts as 100%.
        commitHashes: tests.map(test => ({ hash: test.commit, status: test.status })) // Take all hashes and status' of tests in this group.
      }))
      .map(groupData => [
        groupData.date,
        groupData.flakeRate,
        `<div class="py-2 ps-2">
          <b>${groupData.date.toString()}</b><br>
          <b>Flake Percentage:</b> ${groupData.flakeRate.toFixed(2)}%<br>
          <b>Hashes:</b><br>
          ${groupData.commitHashes.map(({ hash, status }) => `  - ${hash} (${status})`).join("<br>")}
        </div>`
      ])
  );

  const options = {
    title: `Flake Rate by day of ${desiredTest} on ${desiredEnvironment}`,
    width: 900,
    height: 500,
    pointSize: 10,
    pointShape: "circle",
    vAxis: { minValue: 0, maxValue: 1 },
    tooltip: { trigger: "selection", isHtml: true }
  };
  const chart = new google.visualization.LineChart(document.getElementById('chart_div'));
  chart.draw(data, options);
}

init();
