
function displayError(message) {
  console.error(message);
}

async function init() {
  const response = await fetch("content.txt");
  if (!response.ok) {
    const responseText = await response.text();
    displayError(`Failed to fetch data from GCS bucket. Error: ${responseText}`);
    return;
  }

  const responseText = await response.text();
  console.log(responseText);
}

init();
