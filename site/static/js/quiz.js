function selectQuizOption(selectedId) {
    const currentLevel = selectedId.split('/').length - 1;
    $('.option-row').each(function (i) {
        const rowId = $(this).attr('data-quiz-id');
        // don't hide option rows if it has a lower level
        // e.g. when clicking "x86_64" under Linux, we don't want to hide the operating system row
        if ($(this).attr('data-level') < currentLevel) {
            return;
        }
        if (rowId === selectedId) {
            $(this).removeClass('hide');
            $(this).find('.option-button').removeClass('active');
            return;
        }
        // hide all other option rows
        $(this).addClass('hide');
    });
    // hide other answers
    $('.quiz-instruction').addClass('hide');
    // show the selected answer
    $('.quiz-instruction[data-quiz-id=\'' + selectedId + '\']').removeClass('hide');

    const buttons = $('.option-row[data-quiz-id=\'' + selectedId + '\']').find('.option-button');
    // auto-select the first option for the user, to reduce the number of clicks
    if (buttons.length > 0) {
        const btn = buttons.first();
        btn.addClass('active');
        selectQuizOption(btn.attr('data-quiz-id'));
    }
}

function initQuiz() {
    try {
        $('.option-button').click(function(e) {
            $(this).parent().find('.option-button').removeClass('active');
            $(this).addClass('active');
            const dataContainerId = $(this).attr('data-quiz-id');

            selectQuizOption(dataContainerId);
        });
        const userOS = getQuizUserOS();
        // auto-select the OS for user
        const btn = $('.option-button[data-quiz-id=\'/' + userOS + '\']').first();
        btn.addClass('active');
        selectQuizOption(btn.attr('data-quiz-id'));
    } catch(e) {
        const elements = document.getElementsByClassName("quiz-instruction");
        for (let element of elements) {
            element.classList.remove("hide");
        }
    }
}

const getQuizUserOS = () => {
    let os = ['Linux', 'Mac', 'Windows'];
    let userAgent = navigator.userAgent;
    for (let currentOS of os) {
        if (userAgent.indexOf(currentOS) !== -1) {
            if (currentOS === 'Mac') {
                // return the official OS name "macOS" instead
                return 'macOS';
            }
            return currentOS;
        }
    }
    return 'Linux';
}
