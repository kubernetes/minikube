
// Displays an error message to the UI. Any previous message will be erased.
function displayError(message) {
  // Clear the body of all children.
  while (document.body.firstChild) {
    document.body.removeChild(document.body.firstChild);
  }
  const element = document.createElement("p");
  element.innerText = "Error: " + message;
  element.style.color = "red";
  element.style.fontFamily = "Arial";
  element.style.fontWeight = "bold";
  element.style.margin = "5rem";
  document.body.appendChild(element);
}

// Reads `response` into an array of lines while calling `updateProgress` in between.
async function getBodyLinesWithProgress(response, updateProgress) {
  const utf8Decoder = new TextDecoder('utf-8');
  const reader = response.body.getReader();

  const lines = [];
  let pendingText = "";
  let readerDone = false;
  while (!readerDone) {
    // Read a chunk.
    const { value: chunk, done } = await reader.read();
    readerDone = done;
    if (!chunk) {
      continue;
    }
    // Notify the listener of progress.
    updateProgress(chunk.length);
    const decodedChunk = utf8Decoder.decode(chunk);

    const sublines = decodedChunk.split('\n');
    for (let i = 0; i < sublines.length - 1; i++) {
      const fullLine = pendingText + sublines[i];
      pendingText = "";
      if (fullLine !== "") {
        lines.push(fullLine);
      }
    }
    pendingText = sublines[sublines.length - 1];
  }

  // If there is any text remaining, append it.
  if (pendingText !== "") {
    lines.push(pendingText);
  }
  return lines;
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

async function loadTestData(period) {
  const file = period === "last90" ? "data-last-90.csv" : "data.csv";
  const response = await fetch(file, {
    headers: {
      "Cache-Control": "max-age=3600,must-revalidate",
    }
  });
  if (!response.ok) {
    const responseText = await response.text();
    throw `Failed to fetch data from GCS bucket. Error: ${responseText}`;
  }

  const responseDate = new Date(response.headers.get("date").toString());

  const box = document.createElement("div");
  box.style.width = "100%";
  const innerBox = document.createElement("div");
  innerBox.style.margin = "5rem";
  box.appendChild(innerBox);
  const progressBarPrompt = document.createElement("h1");
  progressBarPrompt.style.fontFamily = "Arial";
  progressBarPrompt.style.textAlign = "center";
  progressBarPrompt.innerText = "Downloading data...";
  innerBox.appendChild(progressBarPrompt);
  const progressBar = document.createElement("progress");
  progressBar.setAttribute("max", Number(response.headers.get('Content-Length')));
  progressBar.style.width = "100%";
  innerBox.appendChild(progressBar);
  document.body.appendChild(box);

  let readBytes = 0;
  const lines = await getBodyLinesWithProgress(response, value => {
    readBytes += value;
    progressBar.setAttribute("value", readBytes);
  });
  // Consume the header to ensure the data has the right number of fields.
  const header = lines[0];
  if (header.split(",").length != 9) {
    document.body.removeChild(box);
    throw `Fetched CSV data contains wrong number of fields. Expected: 9. Actual Header: "${header}"`;
  }

  progressBarPrompt.textContent = "Parsing data...";
  progressBar.setAttribute("max", lines.length);

  const testData = [];
  let lineData = ["", "", "", "", "", "", "", "", ""];
  for (let i = 1; i < lines.length; i++) {
    if (i % 30000 === 0) {
      await new Promise(resolve => {
        setTimeout(() => {
          progressBar.setAttribute("value", i);
          resolve();
        });
      });
    }
    const line = lines[i];
    let splitLine = line.split(",");
    if (splitLine.length != 9) {
      console.warn(`Found line with wrong number of fields. Actual: ${splitLine.length} Expected: 9. Line: "${line}"`);
      continue;
    }
    splitLine = splitLine.map((value, index) => value === "" ? lineData[index] : value);
    lineData = splitLine;
    if (!isValidEnumValue(testStatus, splitLine[4])) {
      console.warn(`Invalid test status provided. Actual: ${splitLine[4]} Expected: One of ${Object.values(testStatus).join(", ")}`);
      continue;
    }
    // Skip unsafe dates.
    if (splitLine[1] === "0001-01-01") {
      continue;
    }
    testData.push({
      commit: splitLine[0],
      date: new Date(splitLine[1]),
      environment: splitLine[2],
      name: splitLine[3],
      status: splitLine[4],
      duration: Number(splitLine[5]),
      rootJob: splitLine[6],
      testCount: Number(splitLine[7]),
      totalDuration: Number(splitLine[8]),
    });
  }
  document.body.removeChild(box);
  if (testData.length == 0) {
    throw "Fetched CSV data is empty or poorly formatted.";
  }
  return [testData, responseDate];
}

Array.prototype.min = function() {
  return this.reduce((acc, val) => Math.min(acc, val), Number.MAX_VALUE)
}

Array.prototype.max = function() {
  return this.reduce((acc, val) => Math.max(acc, val), -Number.MAX_VALUE)
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

Array.prototype.oneOfEach = function (keyGetter) {
  return Array.from(new Map(this.map(value => [ keyGetter(value), value ])).values());
}

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
      jobs: tests.map(test => ({ // Take all job ids, statuses, and durations of tests in this group.
        id: test.rootJob,
        status: test.status,
        duration: test.duration
      }))
    }));
}

// Takes a set of test runs (all of the same test), and aggregates them into one element per week.
function aggregateWeeklyRuns(testRuns, weekDates) {
  return testRuns
    // Group runs by the date it ran.
    .groupBy(run => weekDates.findRounded(run.date).getTime())
    // Sort by run date, past to future.
    .sort((a, b) => weekDates.findRounded(a[0].date) - weekDates.findRounded(b[0].date))
    // Map each group to all variables need to format the rows.
    .map(tests => ({
      date: weekDates.findRounded(tests[0].date), // Get one of the dates from the tests, and use it to get the rounded time (which will all be the same).
      flakeRate: tests.map(test => test.status === testStatus.FAILED ? 100 : 0).average(), // Compute average of runs where FAILED counts as 100%.
      duration: tests.map(test => test.duration).average(), // Compute average duration of runs.
      jobs: tests.map(test => ({ // Take all job ids, statuses, and durations of tests in this group.
        id: test.rootJob,
        status: test.status,
        duration: test.duration
      }))
    }));
}

const testGopoghLink = (jobId, environment, testName, status) => {
	const passFail = status === 'Passed' ? 'pass' : 'fail';
	return `https://storage.googleapis.com/minikube-builds/logs/master/${jobId}/${environment}.html${testName ? `#${passFail}_${testName}` : ``}`;
}

function displayTestAndEnvironmentChart(testData, testName, environmentName) {
  const testRuns = testData
    // Filter to only contain unskipped runs of the requested test and requested environment.
    .filter(test => test.name === testName && test.environment === environmentName && test.status !== testStatus.SKIPPED);
  const chartsContainer = document.getElementById('chart_div');
  {
    const data = new google.visualization.DataTable();
    data.addColumn('date', 'Date');
    data.addColumn('number', 'Flake Percentage');
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
    data.addColumn('number', 'Duration');
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });

    data.addRows(
      aggregateRuns(testRuns)
        .map(groupData => [
          groupData.date,
          groupData.flakeRate,
          `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
            <b>Date:</b> ${groupData.date.toLocaleString([], {dateStyle: 'medium'})}<br>
            <b>Flake Percentage:</b> ${groupData.flakeRate.toFixed(2)}%<br>
            <b>Jobs:</b><br>
            ${groupData.jobs.map(({ id, status }) => `  - <a href="${testGopoghLink(id, environmentName, testName, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`,
          groupData.duration,
          `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
            <b>Date:</b> ${groupData.date.toLocaleString([], {dateStyle: 'medium'})}<br>
            <b>Average Duration:</b> ${groupData.duration.toFixed(2)}s<br>
            <b>Jobs:</b><br>
            ${groupData.jobs.map(({ id, duration, status }) => `  - <a href="${testGopoghLink(id, environmentName, testName, status)}">${id}</a> (${duration}s)`).join("<br>")}
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
    const flakeRateContainer = document.createElement("div");
    flakeRateContainer.style.width = "100vw";
    flakeRateContainer.style.height = "100vh";
    chartsContainer.appendChild(flakeRateContainer);
    const chart = new google.visualization.LineChart(flakeRateContainer);
    chart.draw(data, options);
  }
  {
    const dates = testRuns.map(run => run.date.getTime());
    const startDate = new Date(dates.min());
    const endDate = new Date(dates.max());
  
    const weekDates = [];
    let currentDate = startDate;
    while (currentDate < endDate) {
      weekDates.push(currentDate);
      currentDate = new Date(currentDate);
      currentDate.setDate(currentDate.getDate() + 7);
    }
    weekDates.push(currentDate);
    weekDates.findRounded = function (value) {
      let index = this.findIndex(v => value < v);
      if (index == 0) {
        return this[index];
      }
      if (index < 0) {
        index = this.length;
      }
      return this[index - 1];
    }

    const data = new google.visualization.DataTable();
    data.addColumn('date', 'Date');
    data.addColumn('number', 'Flake Percentage');
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
    data.addColumn('number', 'Duration');
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });

    data.addRows(
      aggregateWeeklyRuns(testRuns, weekDates)
        .map(groupData => [
          groupData.date,
          groupData.flakeRate,
          `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
            <b>Date:</b> ${groupData.date.toLocaleString([], {dateStyle: 'medium'})}<br>
            <b>Flake Percentage:</b> ${groupData.flakeRate.toFixed(2)}%<br>
            <b>Jobs:</b><br>
            ${groupData.jobs.map(({ id, status }) => `  - <a href="${testGopoghLink(id, environmentName, testName, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`,
          groupData.duration,
          `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
            <b>Date:</b> ${groupData.date.toLocaleString([], {dateStyle: 'medium'})}<br>
            <b>Average Duration:</b> ${groupData.duration.toFixed(2)}s<br>
            <b>Jobs:</b><br>
            ${groupData.jobs.map(({ id, duration, status }) => `  - <a href="${testGopoghLink(id, environmentName, testName, status)}">${id}</a> (${duration}s)`).join("<br>")}
          </div>`,
        ])
    );

    const options = {
      title: `Flake rate and duration by week of ${testName} on ${environmentName}`,
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
    const flakeRateContainer = document.createElement("div");
    flakeRateContainer.style.width = "100vw";
    flakeRateContainer.style.height = "100vh";
    chartsContainer.appendChild(flakeRateContainer);
    const chart = new google.visualization.LineChart(flakeRateContainer);
    chart.draw(data, options);
  }
}

function createRecentFlakePercentageTable(recentFlakePercentage, previousFlakePercentageMap, environmentName, period) {
  const createCell = (elementType, text) => {
    const element = document.createElement(elementType);
    element.innerHTML = text;
    return element;
  }

  const table = document.createElement("table");
  const tableHeaderRow = document.createElement("tr");
  tableHeaderRow.appendChild(createCell("th", "Rank"));
  tableHeaderRow.appendChild(createCell("th", "Test Name")).style.textAlign = "left";
  tableHeaderRow.appendChild(createCell("th", "Recent Flake Percentage"));
  tableHeaderRow.appendChild(createCell("th", "Growth (since last 15 days)"));
  table.appendChild(tableHeaderRow);
  const tableBody = document.createElement("tbody");
  for (let i = 0; i < recentFlakePercentage.length; i++) {
    const {testName, flakeRate} = recentFlakePercentage[i];
    const row = document.createElement("tr");
    row.appendChild(createCell("td", "" + (i + 1))).style.textAlign = "center";
    row.appendChild(createCell("td", `<a href="${window.location.pathname}?env=${environmentName}&test=${testName}${period === 'last90' ? '&period=last90' : ''}">${testName}</a>`));
    row.appendChild(createCell("td", `${flakeRate.toFixed(2)}%`)).style.textAlign = "right";
    const growth = previousFlakePercentageMap.has(testName) ?
      flakeRate - previousFlakePercentageMap.get(testName) : 0;
    row.appendChild(createCell("td", `<span style="color: ${growth === 0 ? "black" : (growth > 0 ? "red" : "green")}">${growth > 0 ? '+' + growth.toFixed(2) : growth.toFixed(2)}%</span>`));
    tableBody.appendChild(row);
  }
  table.appendChild(tableBody);
  new Tablesort(table);
  return table;
}

function displayEnvironmentChart(testData, environmentName, period) {
  // Number of days to use to look for "flaky-est" tests.
  const dateRange = 15;
  // Number of tests to display in chart.
  const topFlakes = 10;

  testData = testData
    // Filter to only contain unskipped runs of the requested environment.
    .filter(test => test.environment === environmentName && test.status !== testStatus.SKIPPED);

  const testRuns = testData
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
  const previousDates = orderedDates.slice(-2 * dateRange, -dateRange);

  const computeFlakePercentage = (runs, dates) => Array.from(runs).map(([testName, data]) => {
    const {flakeCount, totalCount} = dates.map(date => {
      const dateInfo = data.get(date);
      return dateInfo === undefined ? null : {
        flakeRate: dateInfo.flakeRate,
        runs: dateInfo.jobs.length
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
  
  const recentFlakePercentage = computeFlakePercentage(aggregatedRuns, recentDates)
    .sort((a, b) => b.flakeRate - a.flakeRate);
  const previousFlakePercentageMap = new Map(
    computeFlakePercentage(aggregatedRuns, previousDates)
      .map(({testName, flakeRate}) => [testName, flakeRate]));

  const recentTopFlakes = recentFlakePercentage
    .slice(0, topFlakes)
    .map(({testName}) => testName);

  const chartsContainer = document.getElementById('chart_div');
  {
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
            <b style="display: block">${name}</b><br>
            <b>Date:</b> ${data.date.toLocaleString([], {dateStyle: 'medium'})}<br>
            <b>Flake Percentage:</b> ${data.flakeRate.toFixed(2)}%<br>
            <b>Jobs:</b><br>
            ${data.jobs.map(({ id, status }) => `  - <a href="${testGopoghLink(id, environmentName, name, status)}">${id}</a> (${status})`).join("<br>")}
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
    const flakeRateContainer = document.createElement("div");
    flakeRateContainer.style.width = "100vw";
    flakeRateContainer.style.height = "100vh";
    chartsContainer.appendChild(flakeRateContainer);
    const chart = new google.visualization.LineChart(flakeRateContainer);
    chart.draw(data, options);
  }
  {
    const dates = testData.map(run => run.date.getTime());
    const startDate = new Date(dates.min());
    const endDate = new Date(dates.max());
  
    const weekDates = [];
    let currentDate = startDate;
    while (currentDate < endDate) {
      weekDates.push(currentDate);
      currentDate = new Date(currentDate);
      currentDate.setDate(currentDate.getDate() + 7);
    }
    weekDates.push(currentDate);
    weekDates.findRounded = function (value) {
      let index = this.findIndex(v => value < v);
      if (index == 0) {
        return this[index];
      }
      if (index < 0) {
        index = this.length;
      }
      return this[index - 1];
    }
    const aggregatedWeeklyRuns = new Map(testRuns.map(test => [
      test[0].name,
      new Map(aggregateWeeklyRuns(test, weekDates)
        .map(runDate => [ weekDates.findRounded(runDate.date).getTime(), runDate ]))]));
    const uniqueWeekDates = new Set();
    for (const [_, runDateMap] of aggregatedWeeklyRuns) {
      for (const [dateTime, _] of runDateMap) {
        uniqueWeekDates.add(dateTime);
      }
    }
    const orderedWeekDates = Array.from(uniqueWeekDates).sort();
    const recentWeeklyTopFlakes = computeFlakePercentage(aggregatedWeeklyRuns, [orderedWeekDates[orderedWeekDates.length - 1]])
      .sort((a, b) => b.flakeRate - a.flakeRate)
      .slice(0, topFlakes)
      .map(({testName}) => testName);
    const data = new google.visualization.DataTable();
    data.addColumn('date', 'Date');
    for (const name of recentWeeklyTopFlakes) {
      data.addColumn('number', `Flake Percentage - ${name}`);
      data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
    }
    data.addRows(
      orderedWeekDates.map(dateTime => [new Date(dateTime)].concat(recentTopFlakes.map(name => {
        const data = aggregatedWeeklyRuns.get(name).get(dateTime);
        return data !== undefined ? [
          data.flakeRate,
          `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
            <b style="display: block">${name}</b><br>
            <b>Date:</b> ${data.date.toLocaleString([], {dateStyle: 'medium'})}<br>
            <b>Flake Percentage:</b> ${data.flakeRate.toFixed(2)}%<br>
            <b>Jobs:</b><br>
            ${data.jobs.map(({ id, status }) => `  - <a href="${testGopoghLink(id, environmentName, name, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`
        ] : [null, null];
      })).flat())
    );
    const options = {
      title: `Flake rate by week of top ${topFlakes} of recent test flakiness (past week) on ${environmentName}`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      vAxes: {
        0: { title: "Flake rate", minValue: 0, maxValue: 100 },
      },
      tooltip: { trigger: "selection", isHtml: true }
    };
    const flakeRateContainer = document.createElement("div");
    flakeRateContainer.style.width = "100vw";
    flakeRateContainer.style.height = "100vh";
    chartsContainer.appendChild(flakeRateContainer);
    const chart = new google.visualization.LineChart(flakeRateContainer);
    chart.draw(data, options);
  }
  {
    const jobData = 
      testData
        .groupBy(run => run.date.getTime())
        .map(runDate => ({
          date: runDate[0].date,
          runInfo: runDate
            .oneOfEach(run => run.rootJob)
            .map(run => ({
              commit: run.commit,
              rootJob: run.rootJob,
              testCount: run.testCount,
              totalDuration: run.totalDuration
            }))
        }))
        .sort((a, b) => a.date - b.date)
        .map(({date, runInfo}) => ({
          date,
          runInfo,
          testCount: runInfo.map(job => job.testCount).average(),
          totalDuration: runInfo.map(job => job.totalDuration).average(),
        }));

    const data = new google.visualization.DataTable();
    data.addColumn('date', 'Date');
    data.addColumn('number', 'Test Count');
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
    data.addColumn('number', 'Duration');
    data.addColumn({ type: 'string', role: 'tooltip', 'p': { 'html': true } });
    data.addRows(
      jobData.map(dateInfo => [
        dateInfo.date,
        dateInfo.testCount,
        `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>Date:</b> ${dateInfo.date.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Test Count (averaged): </b> ${+dateInfo.testCount.toFixed(2)}<br>
          <b>Jobs:</b><br>
          ${dateInfo.runInfo.map(job => `  - <a href="${testGopoghLink(job.rootJob, environmentName)}">${job.rootJob}</a> Test count: ${job.testCount}`).join("<br>")}
        </div>`,
        dateInfo.totalDuration,
        `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>Date:</b> ${dateInfo.date.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Total Duration (averaged): </b> ${+dateInfo.totalDuration.toFixed(2)}<br>
          <b>Jobs:</b><br>
          ${dateInfo.runInfo.map(job => `  - <a href="${testGopoghLink(job.rootJob, environmentName)}">${job.rootJob}</a> Total Duration: ${+job.totalDuration.toFixed(2)}s`).join("<br>")}
        </div>`,
      ]));
    const options = {
      title: `Test count and total duration by day on ${environmentName}`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      series: {
        0: { targetAxisIndex: 0 },
        1: { targetAxisIndex: 1 },
      },
      vAxes: {
        0: { title: "Test Count", minValue: 0 },
        1: { title: "Duration (seconds)", minValue: 0 },
      },
      tooltip: { trigger: "selection", isHtml: true }
    };
    const testCountContainer = document.createElement("div");
    testCountContainer.style.width = "100vw";
    testCountContainer.style.height = "100vh";
    chartsContainer.appendChild(testCountContainer);
    const chart = new google.visualization.LineChart(testCountContainer);
    chart.draw(data, options);
  }

  chartsContainer.appendChild(
    createRecentFlakePercentageTable(
      recentFlakePercentage,
      previousFlakePercentageMap,
      environmentName,
      period));
}

async function init() {
  const query = parseUrlQuery(window.location.search);
  const desiredTest = query.test, desiredEnvironment = query.env || "", desiredPeriod = query.period || "";

  google.charts.load('current', { 'packages': ['corechart'] });
  let testData, responseDate;
  try {
    // Wait for Google Charts to load, and for test data to load.
    // Only store the test data (at index 1) into `testData`.
    [testData, responseDate] = (await Promise.all([
      new Promise(resolve => google.charts.setOnLoadCallback(resolve)),
      loadTestData(desiredPeriod)
    ]))[1];
  } catch (err) {
    displayError(err);
    return;
  }

  if (desiredTest === undefined) {
    displayEnvironmentChart(testData, desiredEnvironment, desiredPeriod);
  } else {
    displayTestAndEnvironmentChart(testData, desiredTest, desiredEnvironment);
  }
  document.querySelector('#data_date_container').style.display = 'block';
  document.querySelector('#data_date').innerText = responseDate.toLocaleString();
  let periodDisplay, newURL;

  // we're going to take the current page URL (desiredPeriod) and modify it to create the link to the other page
  if (desiredPeriod === 'last90') {
    // remove '&period=last90' to make a link to the all-time data page
    otherPeriodURL = window.location.href.replace(/&?period=last90/gi, '');
    periodDisplay = `Currently viewing last 90 days of data: <a href="` + otherPeriodURL + `">View all-time data</a>`;
  } else {
    // add '&period=last90' to make a link to the last 90 days page
    otherPeriodURL = window.location.href + '&period=last90';
    periodDisplay = `Currently viewing all-time data: <a href="` + otherPeriodURL + `">View last 90 days of data</a>`;
  }
  document.querySelector('#period_display').innerHTML = periodDisplay;
}

init();
