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

const testGopoghLink = (jobId, environment, testName, status) => {
  return `https://storage.googleapis.com/minikube-builds/logs/master/${jobId}/${environment}.html${testName ? `#${status}_${testName}` : ``}`;
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

function createRecentNumberOfFailTable(summaryTable) {
  const createCell = (elementType, text) => {
      const element = document.createElement(elementType);
      element.innerHTML = text;
      return element;
  }
  const table = document.createElement("table");
  const tableHeaderRow = document.createElement("tr");
  tableHeaderRow.appendChild(createCell("th", "Rank"));
  tableHeaderRow.appendChild(createCell("th", "Env Name")).style.textAlign = "left";
  tableHeaderRow.appendChild(createCell("th", "Recent Number of Fails"));
  tableHeaderRow.appendChild(createCell("th", "Growth (since last 15 days)"));
  table.appendChild(tableHeaderRow);
  const tableBody = document.createElement("tbody");
  for (let i = 0; i < summaryTable.length; i++) {
      const {
          envName,
          recentNumberOfFail,
          growth
      } = summaryTable[i];
      const row = document.createElement("tr");
      row.appendChild(createCell("td", "" + (i + 1))).style.textAlign = "center";
      row.appendChild(createCell("td", `<a href="${window.location.pathname}?env=${envName}">${envName}</a>`));
      row.appendChild(createCell("td", recentNumberOfFail)).style.textAlign = "right";
      row.appendChild(createCell("td", `<span style="color: ${growth === 0 ? "black" : (growth > 0 ? "red" : "green")}">${growth > 0 ? '+' + growth : growth}</span>`));
      tableBody.appendChild(row);
  }
  table.appendChild(tableBody);
  new Tablesort(table);
  return table;
}


function createRecentFlakePercentageTable(recentFlakePercentTable, query) {
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
  for (let i = 0; i < recentFlakePercentTable.length; i++) {
      const {
          testName,
          recentFlakePercentage,
          growthRate
      } = recentFlakePercentTable[i];
      const row = document.createElement("tr");
      row.appendChild(createCell("td", "" + (i + 1))).style.textAlign = "center";
      row.appendChild(createCell("td", `<a href="${window.location.pathname}?env=${query.env}&test=${testName}">${testName}</a>`));
      row.appendChild(createCell("td", recentFlakePercentage + "%")).style.textAlign = "right";
      row.appendChild(createCell("td", `<span style="color: ${growthRate === 0 ? "black" : (growthRate > 0 ? "red" : "green")}">${growthRate > 0 ? '+' + growthRate : growthRate}%</span>`));
      tableBody.appendChild(row);
  }
  table.appendChild(tableBody);
  new Tablesort(table);
  return table;
}

function displayTestAndEnvironmentChart(data, query) {
  const chartsContainer = document.getElementById('chart_div');

  const dayData = data.flakeByDay
  const dayChart = new google.visualization.DataTable();
  dayChart.addColumn('date', 'Date');
  dayChart.addColumn('number', 'Flake Percentage');
  dayChart.addColumn({
      type: 'string',
      role: 'tooltip',
      'p': {
          'html': true
      }
  });
  dayChart.addColumn('number', 'Duration');
  dayChart.addColumn({
      type: 'string',
      role: 'tooltip',
      'p': {
          'html': true
      }
  });

  dayChart.addRows(
      dayData
      .map(groupData => {
          let dataArr = groupData.commitResultsAndDurations.split(',')
          dataArr = dataArr.map((commit) => commit.split(":"))
          const resultArr = dataArr.map((commit) => ({
              id: commit[commit.length - 3],
              status: (commit[commit.length - 2]).trim()
          }))
          const durationArr = dataArr.map((commit) => ({
              id: commit[commit.length - 3],
              status: (commit[commit.length - 2]).trim(),
              duration: (commit[commit.length - 1]).trim()
          }))

          return [
              new Date(groupData.startOfDate),
              groupData.flakePercentage,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>Date:</b> ${groupData.startOfDate.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Flake Percentage:</b> ${groupData.flakePercentage.toFixed(2)}%<br>
          <b>Jobs:</b><br>
          ${resultArr.map(({ id, status }) => `  - <a href="${testGopoghLink(id, query.env, query.test, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`,
              groupData.avgDuration,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>Date:</b> ${groupData.startOfDate.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Average Duration:</b> ${groupData.avgDuration.toFixed(2)}s<br>
          <b>Jobs:</b><br>
          ${durationArr.map(({ id, duration, status }) => `  - <a href="${testGopoghLink(id, query.env, query.test, status)}">${id}</a> (${duration}s)`).join("<br>")}
          </div>`,
          ]
      })
  );
  const dayOptions = {
      title: `Flake rate and duration by day of ${query.test} on ${query.env}`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      series: {
          0: {
              targetAxisIndex: 0
          },
          1: {
              targetAxisIndex: 1
          },
      },
      vAxes: {
          0: {
              title: "Flake rate",
              minValue: 0,
              maxValue: 100
          },
          1: {
              title: "Duration (seconds)"
          },
      },
      colors: ['#dc3912', '#3366cc'],
      tooltip: {
          trigger: "selection",
          isHtml: true
      }
  };
  const flakeRateDayContainer = document.createElement("div");
  flakeRateDayContainer.style.width = "100vw";
  flakeRateDayContainer.style.height = "100vh";
  chartsContainer.appendChild(flakeRateDayContainer);
  const dChart = new google.visualization.LineChart(flakeRateDayContainer);
  dChart.draw(dayChart, dayOptions);

  const weekData = data.flakeByWeek
  const weekChart = new google.visualization.DataTable();
  weekChart.addColumn('date', 'Date');
  weekChart.addColumn('number', 'Flake Percentage');
  weekChart.addColumn({
      type: 'string',
      role: 'tooltip',
      'p': {
          'html': true
      }
  });
  weekChart.addColumn('number', 'Duration');
  weekChart.addColumn({
      type: 'string',
      role: 'tooltip',
      'p': {
          'html': true
      }
  });

  console.log(weekChart)
  weekChart.addRows(
      weekData
      .map(groupData => {
          let dataArr = groupData.commitResultsAndDurations.split(',')
          dataArr = dataArr.map((commit) => commit.split(":"))
          const resultArr = dataArr.map((commit) => ({
              id: commit[commit.length - 3],
              status: (commit[commit.length - 2]).trim()
          }))
          const durationArr = dataArr.map((commit) => ({
              id: commit[commit.length - 3],
              status: (commit[commit.length - 2]).trim(),
              duration: (commit[commit.length - 1]).trim()
          }))

          return [
              new Date(groupData.startOfDate),
              groupData.flakePercentage,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>Date:</b> ${groupData.startOfDate.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Flake Percentage:</b> ${groupData.flakePercentage.toFixed(2)}%<br>
          <b>Jobs:</b><br>
          ${resultArr.map(({ id, status }) => `  - <a href="${testGopoghLink(id, query.env, query.test, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`,
              groupData.avgDuration,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b>Date:</b> ${groupData.startOfDate.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Average Duration:</b> ${groupData.avgDuration.toFixed(2)}s<br>
          <b>Jobs:</b><br>
          ${durationArr.map(({ id, duration, status }) => `  - <a href="${testGopoghLink(id, query.env, query.test, status)}">${id}</a> (${duration}s)`).join("<br>")}
          </div>`,
          ]
      })
  );
  const weekOptions = {
      title: `Flake rate and duration by week of ${query.test} on ${query.env}`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      series: {
          0: {
              targetAxisIndex: 0
          },
          1: {
              targetAxisIndex: 1
          },
      },
      vAxes: {
          0: {
              title: "Flake rate",
              minValue: 0,
              maxValue: 100
          },
          1: {
              title: "Duration (seconds)"
          },
      },
      colors: ['#dc3912', '#3366cc'],
      tooltip: {
          trigger: "selection",
          isHtml: true
      }
  };
  const flakeRateWeekContainer = document.createElement("div");
  flakeRateWeekContainer.style.width = "100vw";
  flakeRateWeekContainer.style.height = "100vh";
  chartsContainer.appendChild(flakeRateWeekContainer);
  const wChart = new google.visualization.LineChart(flakeRateWeekContainer);
  wChart.draw(weekChart, weekOptions);
}

function displaySummaryChart(data) {
  const chartsContainer = document.getElementById('chart_div');
  const summaryData = data.summaryAvgFail

  const uniqueDayDates = new Set();
  const summaryEnvDateMap = {};
  for (const envDay of summaryData) {
      const {
          startOfDate,
          envName,
          avgFailedTests,
          avgDuration
      } = envDay
      uniqueDayDates.add(startOfDate)
      if (!summaryEnvDateMap[envName]) {
          summaryEnvDateMap[envName] = {};
      }
      summaryEnvDateMap[envName][startOfDate] = {
          avgFailedTests,
          avgDuration
      }
  }
  const uniqueEnvs = Object.keys(summaryEnvDateMap);
  const orderedDayDates = Array.from(uniqueDayDates).sort()


  const dayChart = new google.visualization.DataTable();
  dayChart.addColumn('date', 'Date');
  for (const env of uniqueEnvs) {
      dayChart.addColumn('number', `${env}`);
      dayChart.addColumn({
          type: 'string',
          role: 'tooltip',
          'p': {
              'html': true
          }
      });
  }
  dayChart.addRows(orderedDayDates.map(dateTime => [new Date(dateTime)].concat(uniqueEnvs.map(name => {
      const avgVal = summaryEnvDateMap[name][dateTime];
      if (avgVal !== undefined) {
          const {
              avgFailedTests,
          } = avgVal
          return [
              avgFailedTests,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b style="display: block">${name}</b><br>
          <b>Date:</b> ${dateTime.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Number of Failed Tests (avg):</b> ${+avgFailedTests.toFixed(2)}<br>
        </div>`
          ]
      }
      return [null, null];
  })).flat()))

  const dayOptions = {
      title: `Average Daily Failed Tests`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      vAxes: {
          0: {
              title: "# of Failed Tests",
              minValue: 0
          },
      },
      tooltip: {
          trigger: "selection",
          isHtml: true
      }
  };
  // Create the chart and draw it
  const summaryDayContainer = document.createElement("div");
  summaryDayContainer.style.width = "100vw";
  summaryDayContainer.style.height = "100vh";
  chartsContainer.appendChild(summaryDayContainer);
  const dChart = new google.visualization.LineChart(summaryDayContainer);
  dChart.draw(dayChart, dayOptions);


  const durChart = new google.visualization.DataTable();
  durChart.addColumn('date', 'Date');
  for (const env of uniqueEnvs) {
      durChart.addColumn('number', `${env}`);
      durChart.addColumn({
          type: 'string',
          role: 'tooltip',
          'p': {
              'html': true
          }
      });
  }
  durChart.addRows(orderedDayDates.map(dateTime => [new Date(dateTime)].concat(uniqueEnvs.map(name => {
      const avgVal = summaryEnvDateMap[name][dateTime];
      if (avgVal !== undefined) {
          const {
              avgDuration
          } = avgVal
          return [
              avgDuration,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b style="display: block">${name}</b><br>
          <b>Date:</b> ${dateTime.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Duration (avg):</b> ${+avgDuration.toFixed(2)}<br>
        </div>`
          ]
      }
      return [null, null];
  })).flat()))

  const durOptions = {
      title: `Average Total Duration per day`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      vAxes: {
          0: {
              title: "Total Duration",
              minValue: 0
          },
      },
      tooltip: {
          trigger: "selection",
          isHtml: true
      }
  };
  // Create the chart and draw it
  const summaryDurContainer = document.createElement("div");
  summaryDurContainer.style.width = "100vw";
  summaryDurContainer.style.height = "100vh";
  chartsContainer.appendChild(summaryDurContainer);
  const durationChart = new google.visualization.LineChart(summaryDurContainer);
  durationChart.draw(durChart, durOptions);


  chartsContainer.appendChild(createRecentNumberOfFailTable(data.summaryTable))
}

function displayEnvironmentChart(data, query) {
  const chartsContainer = document.getElementById('chart_div');

  //By Day Chart

  const dayData = data.flakeRateByDay
  const uniqueDayTestNames = new Set();
  const uniqueDayDates = new Set();
  for (const flakeDay of dayData) {
      uniqueDayTestNames.add(flakeDay.testName);
      uniqueDayDates.add(flakeDay.startOfDate)
  }
  const uniqueDayTestNamesArray = Array.from(uniqueDayTestNames);
  const orderedDayDates = Array.from(uniqueDayDates).sort();
  const flakeDayDataMap = {};
  dayData.forEach((day) => {
      const {
          testName,
          startOfDate,
          flakePercentage,
          commitResults
      } = day;
      // If the test name doesn't exist in the map, create a new entry
      if (!flakeDayDataMap[testName]) {
          flakeDayDataMap[testName] = {};
      }
      // Set the flakePercentage for the corresponding startOfDate
      flakeDayDataMap[testName][startOfDate] = {
          fp: flakePercentage,
          cr: commitResults
      };
  });
  const dayChart = new google.visualization.DataTable();
  dayChart.addColumn('date', 'Date');
  for (const testName of uniqueDayTestNamesArray) {
      dayChart.addColumn('number', `${testName}`);
      dayChart.addColumn({
          type: 'string',
          role: 'tooltip',
          'p': {
              'html': true
          }
      });
  }
  dayChart.addRows(orderedDayDates.map(dateTime => [new Date(dateTime)].concat(uniqueDayTestNamesArray.map(name => {
      const fpAndCr = flakeDayDataMap[name][dateTime];
      if (fpAndCr !== undefined) {
          const {
              fp,
              cr
          } = fpAndCr
          let commitArr = cr.split(",")
          commitArr = commitArr.map((commit) => commit.split(":"))
          commitArr = commitArr.map((commit) => ({
              id: commit[commit.length - 2],
              status: (commit[commit.length - 1]).trim()
          }))
          return [
              fp,
              `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b style="display: block">${name}</b><br>
          <b>Date:</b> ${dateTime.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Flake Percentage:</b> ${+fp.toFixed(2)}%<br>
          <b>Jobs:</b><br>
          ${commitArr.map(({ id, status }) => `  - <a href="${testGopoghLink(id, query.env, name, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`
          ]
      }
      return [null, null];
  })).flat()))

  const dayOptions = {
      title: `Flake rate by day of top ${uniqueDayTestNamesArray.length} recent test flakiness (past 15 days) on ${query.env}`,
      width: window.innerWidth,
      height: window.innerHeight,
      pointSize: 10,
      pointShape: "circle",
      vAxes: {
          0: {
              title: "Flake rate",
              minValue: 0,
              maxValue: 100
          },
      },
      tooltip: {
          trigger: "selection",
          isHtml: true
      }
  };
  // Create the chart and draw it
  const flakeRateDayContainer = document.createElement("div");
  flakeRateDayContainer.style.width = "100vw";
  flakeRateDayContainer.style.height = "100vh";
  chartsContainer.appendChild(flakeRateDayContainer);
  const dChart = new google.visualization.LineChart(flakeRateDayContainer);
  dChart.draw(dayChart, dayOptions);

  // Weekly Chart

  const weekData = data.flakeRateByWeek
  const uniqueWeekTestNames = new Set();
  const uniqueWeekDates = new Set();
  for (const flakeWeek of weekData) {
      uniqueWeekTestNames.add(flakeWeek.testName);
      uniqueWeekDates.add(flakeWeek.startOfDate)
  }
  const uniqueWeekTestNamesArray = Array.from(uniqueWeekTestNames);
  const orderedWeekDates = Array.from(uniqueWeekDates).sort();
  const flakeWeekDataMap = {};
  weekData.forEach((week) => {
      const {
          testName,
          startOfDate,
          flakePercentage,
          commitResults
      } = week;
      // If the test name doesn't exist in the map, create a new entry
      if (!flakeWeekDataMap[testName]) {
          flakeWeekDataMap[testName] = {};
      }
      // Set the flakePercentage for the corresponding startOfDate
      flakeWeekDataMap[testName][startOfDate] = {
          fp: flakePercentage,
          cr: commitResults
      };
  });
  {
      // Create the DataTable
      const weekChart = new google.visualization.DataTable();
      // Add the columns to the DataTable
      weekChart.addColumn('date', 'Date');
      for (const testName of uniqueWeekTestNamesArray) {
          weekChart.addColumn('number', `${testName}`);
          weekChart.addColumn({
              type: 'string',
              role: 'tooltip',
              'p': {
                  'html': true
              }
          });
      }
      weekChart.addRows(orderedWeekDates.map(dateTime => [new Date(dateTime)].concat(uniqueWeekTestNamesArray.map(name => {
          const fpAndcr = flakeWeekDataMap[name][dateTime];
          if (fpAndcr != undefined) {
              const {
                  fp,
                  cr
              } = fpAndcr
              let commitArr = cr.split(",")
              commitArr = commitArr.map((commit) => commit.split(":"))
              commitArr = commitArr.map((commit) => ({
                  id: commit[commit.length - 2],
                  status: (commit[commit.length - 1]).trim()
              }))
              return [
                  fp,
                  `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
          <b style="display: block">${name}</b><br>
          <b>Date:</b> ${dateTime.toLocaleString([], {dateStyle: 'medium'})}<br>
          <b>Flake Percentage:</b> ${+fp.toFixed(2)}%<br>
          <b>Jobs:</b><br>
          ${commitArr.map(({ id, status }) => `  - <a href="${testGopoghLink(id, query.env, name, status)}">${id}</a> (${status})`).join("<br>")}
          </div>`
              ];
          }
          return [null, null];
      })).flat()))

      const weekOptions = {
          title: `Flake rate by week of top ${uniqueWeekTestNamesArray.length} of recent test flakiness (past week) on ${query.env}`,
          width: window.innerWidth,
          height: window.innerHeight,
          pointSize: 10,
          pointShape: "circle",
          vAxes: {
              0: {
                  title: "Flake rate",
                  minValue: 0,
                  maxValue: 100
              },
          },
          tooltip: {
              trigger: "selection",
              isHtml: true
          }
      };
      // Create the chart and draw it
      const flakeRateWeekContainer = document.createElement("div");
      flakeRateWeekContainer.style.width = "100vw";
      flakeRateWeekContainer.style.height = "100vh";
      chartsContainer.appendChild(flakeRateWeekContainer);
      const wChart = new google.visualization.LineChart(flakeRateWeekContainer);
      wChart.draw(weekChart, weekOptions);
  }

  // Duration Chart

  {
      const durationChart = new google.visualization.DataTable();
      const durationData = data.countsAndDurations
      durationChart.addColumn('date', 'Date');
      durationChart.addColumn('number', 'Test Count');
      durationChart.addColumn({
          type: 'string',
          role: 'tooltip',
          'p': {
              'html': true
          }
      });
      durationChart.addColumn('number', 'Duration');
      durationChart.addColumn({
          type: 'string',
          role: 'tooltip',
          'p': {
              'html': true
          }
      });
      durationChart.addRows(
          durationData.map(dateInfo => {
              let countArr = dateInfo.commitCounts.split(",")
              countArr = countArr.map((commit) => commit.split(":"))
              countArr = countArr.map((commit) => ({
                  rootJob: commit[commit.length - 2],
                  testCount: +(commit[commit.length - 1]).trim()
              }))
              let durationArr = dateInfo.commitDurations.split(",")
              durationArr = durationArr.map((commit) => commit.split(":"))
              durationArr = durationArr.map((commit) => ({
                  rootJob: commit[commit.length - 2],
                  totalDuration: +(commit[commit.length - 1]).trim()
              }))
              return [
                  new Date(dateInfo.startOfDate),
                  dateInfo.testCount,
                  `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
        <b>Date:</b> ${dateInfo.startOfDate.toLocaleString([], {dateStyle: 'medium'})}<br>
        <b>Test Count (averaged): </b> ${+dateInfo.testCount.toFixed(2)}<br>
        <b>Jobs:</b><br>
        ${countArr.map(job => `  - <a href="${testGopoghLink(job.rootJob, query.env)}">${job.rootJob}</a> Test count: ${job.testCount}`).join("<br>")}
      </div>`,
                  dateInfo.duration,
                  `<div style="padding: 1rem; font-family: 'Arial'; font-size: 14">
        <b>Date:</b> ${dateInfo.startOfDate.toLocaleString([], {dateStyle: 'medium'})}<br>
        <b>Total Duration (averaged): </b> ${+dateInfo.duration.toFixed(2)}<br>
        <b>Jobs:</b><br>
        ${durationArr.map(job => `  - <a href="${testGopoghLink(job.rootJob, query.env)}">${job.rootJob}</a> Total Duration: ${+job.totalDuration.toFixed(2)}s`).join("<br>")}
      </div>`,
              ]
          }));
      const durOptions = {
          title: `Test count and total duration by day on ${query.env}`,
          width: window.innerWidth,
          height: window.innerHeight,
          pointSize: 10,
          pointShape: "circle",
          series: {
              0: {
                  targetAxisIndex: 0
              },
              1: {
                  targetAxisIndex: 1
              },
          },
          vAxes: {
              0: {
                  title: "Test Count",
                  minValue: 0
              },
              1: {
                  title: "Duration (seconds)",
                  minValue: 0
              },
          },
          tooltip: {
              trigger: "selection",
              isHtml: true
          }
      };
      const envDurationContainer = document.createElement("div");
      envDurationContainer.style.width = "100vw";
      envDurationContainer.style.height = "100vh";
      chartsContainer.appendChild(envDurationContainer);
      const durChart = new google.visualization.LineChart(envDurationContainer);
      durChart.draw(durationChart, durOptions);
  }

  chartsContainer.appendChild(createRecentFlakePercentageTable(data.recentFlakePercentTable, query))
}

function createTopnDropdown(currentTopn) {
  const dropdownContainer = document.createElement("div");
  dropdownContainer.style.margin = "1rem";

  const dropdownLabel = document.createElement("label");
  dropdownLabel.innerText = "Select topn value: ";

  const dropdown = document.createElement("select");
  dropdown.id = "topnDropdown";

  const values = [3, 5, 10, 15];
  values.forEach(value => {
      const option = document.createElement("option");
      option.value = value;
      option.text = value;
      if (value.toString() === currentTopn) {
          option.selected = true;
      }
      dropdown.appendChild(option);
  });

  dropdown.addEventListener("change", () => {
      const selectedValue = dropdown.value;
      const currentURL = new URL(window.location.href);
      currentURL.searchParams.set("tests_in_top", selectedValue);
      window.location.href = currentURL.href;
  });

  dropdownContainer.appendChild(dropdownLabel);
  dropdownContainer.appendChild(dropdown);

  document.getElementById('dropdown_container').appendChild(dropdownContainer)
}

function displayGopoghVersion(verData) {
  const footerElement = document.getElementById('version_div');
  const version = verData.version

  footerElement.className = "mdl-mega-footer";
  footerElement.innerHTML = "generated by <a href=\"https://github.com/medyagh/gopogh/\">Gopogh " + version + "</a>";
}


async function init() {
  const query = parseUrlQuery(window.location.search);
  const desiredTest = query.test,
      desiredEnvironment = query.env,
      desiredPeriod = query.period || "",
      desiredTestNumber = query.tests_in_top || "";
  const currentTopn = query.tests_in_top || "10"; // Default to 10 (for top 10 tests)

  google.charts.load('current', {
      'packages': ['corechart']
  });
  try {
      // Wait for Google Charts to load
      await new Promise(resolve => google.charts.setOnLoadCallback(resolve));

      let url;
      const basePath = 'https://gopogh-server-tts3vkcpgq-uc.a.run.app' // Base Server Path. Modify to actual server path if deploying
      if (desiredEnvironment === undefined) {
          // URL for displaySummaryChart
          url = basePath + '/summary'
      } else if (desiredTest === undefined) {
          // URL for displayEnvironmentChart
          url = basePath + '/env' + '?env=' + desiredEnvironment + '&tests_in_top=' + desiredTestNumber;
      } else {
          // URL for displayTestAndEnvironmentChart
          url = basePath + '/test' + '?env=' + desiredEnvironment + '&test=' + desiredTest;
      }

      // Fetch data from the determined URL
      const response = await fetch(url);
      if (!response.ok) {
          throw new Error('Network response was not ok');
      }
      const data = await response.json();
      console.log(data)

      // Call the appropriate chart display function based on the desired condition
      if (desiredTest == undefined && desiredEnvironment === undefined) {
          displaySummaryChart(data)
      } else if (desiredTest === undefined) {
          createTopnDropdown(currentTopn);
          displayEnvironmentChart(data, query);
      } else {
          displayTestAndEnvironmentChart(data, query);
      }
      url = basePath + '/version'

      const verResponse = await fetch(url);
      if (!verResponse.ok) {
          throw new Error('Network response was not ok');
      }
      const verData = await verResponse.json();
      console.log(verData)
      displayGopoghVersion(verData)
  } catch (err) {
      displayError(err);
  }
}

init();
