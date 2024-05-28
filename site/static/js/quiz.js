function selectQuizOption(selectedId, autoselect = true) {
  const currentLevel = selectedId.split("/").length - 1;
  $(".option-row").each(function (i) {
    const rowId = $(this).attr("data-quiz-id");
    // don't hide option rows if it has a lower level
    // e.g. when clicking "x86_64" under Linux, we don't want to hide the operating system row
    if ($(this).attr("data-level") < currentLevel) {
      return;
    }
    if (rowId === selectedId) {
      $(this).removeClass("hide");
      $(this).find(".option-button").removeClass("active");
      return;
    }
    // hide all other option rows
    $(this).addClass("hide");
  });
  // hide other answers
  $(".quiz-instruction").addClass("hide");
  // show the selected answer
  $(".quiz-instruction[data-quiz-id='" + selectedId + "']").removeClass("hide");

  const buttons = $(".option-row[data-quiz-id='" + selectedId + "']").find(
    ".option-button"
  );
  //auto-select the first option for the user, to reduce the number of clicks
  if (buttons.length > 0) {
    if (autoselect) {
      const btn = buttons.first();
      const dataContainerId = btn.attr("data-quiz-id");
      btn.addClass("active");
      const url = new URL(window.location);
      url.searchParams.set("arch", dataContainerId); // Add selectedId as query parameter

      // Update the browser's location with the new URL
      window.history.replaceState({}, document.title, url);

      selectQuizOption(dataContainerId);
    }
  }
}

function initQuiz() {
  try {
    $(".option-button").click(function (e) {
      $(this).parent().find(".option-button").removeClass("active");
      $(this).addClass("active");
      const dataContainerId = $(this).attr("data-quiz-id");

      const url = new URL(window.location);
      url.searchParams.set("arch", dataContainerId); // Add selectedId as query parameter

      window.history.replaceState({}, document.title, url);
      // Update the browser's location with the new URL

      selectQuizOption(dataContainerId);
    });
    let userOS = getUserOS().toLowerCase();
    if (userOS === "Mac") {
      // use the name "macos" to match the button
      userOS = "macos";
    }
    $(".option-row[data-level=0]").removeClass("hide");

    const urlParams = new URLSearchParams(window.location.search);
    const archParam = urlParams.get("arch");

    //checks for query params and process each option one by one

    if (archParam) {
      const options = archParam.split("/").filter(Boolean);
      let quizId = "";
      options.forEach((option, index) => {
        quizId = quizId + "/" + option;
        const archBtn = $(
          ".option-button[data-quiz-id='" + quizId + "']"
        ).first();
        archBtn.addClass("active");

        // passes false as argument when there are more options to process to prevent auto selection of 1st option in following options
        if (index === option.length - 1) {
          selectQuizOption(archBtn.attr("data-quiz-id"));
        } else {
          selectQuizOption(archBtn.attr("data-quiz-id"), false);
        }
      });
    } else {
      // auto-select the OS for user
      const btn = $(".option-button[data-quiz-id='/" + userOS + "']").first();
      btn.addClass("active");
      selectQuizOption(btn.attr("data-quiz-id"));
    }
  } catch (e) {
    const elements = document.getElementsByClassName("quiz-instruction");
    for (let element of elements) {
      element.classList.remove("hide");
    }
  }
}
