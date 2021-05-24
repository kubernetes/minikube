
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

async function init() {
  const response = await fetch("data.csv");
  if (!response.ok) {
    const responseText = await response.text();
    displayError(`Failed to fetch data from GCS bucket. Error: ${responseText}`);
    return;
  }

  const lines = bodyByLinesIterator(response);
  // Consume the header to ensure the data has the right number of fields.
  const header = (await lines.next()).value;
  if (header.split(",").length != 5) {
    displayError(`Fetched CSV data contains wrong number of fields. Expected: 5. Actual Header: "${header}"`);
    return;
  }

  const testData = [];
  for await (const line of lines) {
    const splitLine = line.split(",");
    if (splitLine.length != 5) {
      console.warn(`Found line with wrong number of fields. Actual: ${splitLine.length} Expected: 5. Line: "${line}"`);
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
      status: splitLine[4]
    });
  }
  if (testData.length == 0) {
    displayError("Fetched CSV data is empty or poorly formatted.");
    return;
  }

  console.log(testData);
}

init();
