/* Tabs JS implementation. Borrowed from Skaffold */
function initTabs() {
  try {
    $(".tab-content")
      .children(".tab-pane")
      .each(function (idx, item) {
        var navTabs = $(this).closest(".code-tabs").children(".nav-tabs"),
          title = escape($(this).attr("title")).replace(/%20/g, " "),
          os = escape($(this).attr("os") || "");
        navTabs.append(
          '<li class="nav-tab ' +
            os +
            '"><a href="#' +
            title +
            '" class="nav-tab">' +
            title +
            "</a></li>"
        );
      });

    $(".code-tabs ul.nav-tabs").each(function () {
      let tabSelector = getTabSelector(this);
      $(this)
        .find("li" + tabSelector)
        .addClass("active");
    });

    $(".code-tabs .tab-content").each(function () {
      let tabSelector = getTabSelector(this);
      $(this)
        .find("div" + tabSelector)
        .addClass("active");
    });

    $(".nav-tabs a").click(function (e) {
      e.preventDefault();
      var tab = $(this).parent(),
        tabIndex = tab.index(),
        tabPanel = $(this).closest(".code-tabs"),
        tabPane = tabPanel
          .find(".tab-content:first")
          .children(".tab-pane")
          .eq(tabIndex);
      tab.siblings().removeClass("active");
      tabPane.siblings().removeClass("active");
      tab.addClass("active");
      tabPane.addClass("active");

      // changes the anchor in the url
      var tabTitle = $(this).attr("href");
      window.location.hash = tabTitle;
    });

    const hash = window.location.hash;

    // checks for anchor in the url and simulate anchor click to see the particular tab
    if (hash) {
      const tabTitle = unescape(hash.replace("#", ""));
      const tab = $(".nav-tabs a[href='#" + tabTitle + "']");
      tab.click(); // Trigger click to activate the tab
    }
    const url = new URL(window.location);
    if (url.pathname === "/docs/handbook/addons/ingress-dns/") {
      url.hash = getUserOS();
      window.history.replaceState({}, document.title, url);
    }
  } catch (e) {
    const elements = document.getElementsByClassName("tab-pane");
    for (let element of elements) {
      element.style.display = "block";
      const title = document.createElement("h3");
      title.innerText = element.title;
      title.classList.add("backup-tab-title");
      element.prepend(title);
    }
  }
}

const getTabSelector = (currElement) => {
  let osSelector = "." + getUserOS();
  let hasMatchingOSTab = $(currElement).find(osSelector).length;
  return hasMatchingOSTab ? osSelector : ":first";
};

const getUserOS = () => {
  let os = ["Linux", "Mac", "Windows"];
  let userAgent = navigator.userAgent;
  for (let currentOS of os) {
    if (userAgent.indexOf(currentOS) !== -1) return currentOS;
  }
  return "Linux";
};
