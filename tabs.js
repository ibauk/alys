"use strict";
function getFirstChildWithTagName(element, tagName) {
  // Tabbing
  for (let i = 0; i < element.childNodes.length; i++) {
    if (element.childNodes[i].nodeName == tagName) return element.childNodes[i];
  }
}

let tabLinks = [];
let contentDivs = [];

function setupTabs() {
  let hasTabs = document.getElementById("tabs");
  if (!hasTabs) return;

  let tabListItems = hasTabs.childNodes;

  let tab2show = 0;
  let vartab2show = document.getElementById("tab2show");
  if (vartab2show) {
    tab2show = vartab2show.value;
    if (tab2show >= tabListItems.length) tab2show = 0;
  }

  for (let i = 0; i < tabListItems.length; i++) {
    if (tabListItems[i].nodeName == "LI") {
      let tabLink = getFirstChildWithTagName(tabListItems[i], "A");
      let id = tabsGetHash(tabLink.getAttribute("href"));
      tabLinks[id] = tabLink;
      contentDivs[id] = document.getElementById(id);
    }
  }

  let i = 0;

  for (let id in tabLinks) {
    tabLinks[id].onclick = tabsShowTab;
    tabLinks[id].onfocus = function () {
      this.blur();
    };
    if (i == tab2show) tabLinks[id].className = "selected";
    i++;
  }

  // Hide all content divs except the first
  i = 0;

  for (let id in contentDivs) {
    if (i != tab2show) {
      contentDivs[id].classList.remove("tabContent");
      contentDivs[id].classList.add("tabContenthide");
    }
    let legend = getFirstChildWithTagName(contentDivs[id], "LEGEND");

    if (legend) legend.innerText = "";
    i++;
  }
}

function tabsGetHash(url) {
  // Tabbing
  var hashPos = url.lastIndexOf("#");
  return url.substring(hashPos + 1);
}

function tabsShowTab() {
  let tab2show = 0;
  let vartab2show = document.getElementById("tab2show");

  let selectedId = tabsGetHash(this.getAttribute("href"));

  // Highlight the selected tab, and dim all others.
  // Also show the selected content div, and hide all others.
  for (let id in contentDivs) {
    if (id == selectedId) {
      tabLinks[id].className = "selected";
      contentDivs[id].classList.remove("tabContenthide");
      contentDivs[id].classList.add("tabContent");
      if (vartab2show) {
        vartab2show.value = tab2show;
        let links = document.getElementsByClassName("navLink");
        for (let l = 0; l < links.length; l++) {
          console.log("l==" + links[l].getAttribute("href"));
          let p = /tab=\d&/;
          let r = p.exec(links[l].href);
          if (r) {
            links[l].href = links[l].href.replace(p, "tab=" + tab2show + "&");
          }
        }
      }
    } else {
      tabLinks[id].className = "";
      contentDivs[id].classList.remove("tabContent");
      contentDivs[id].classList.add("tabContenthide");
    }
    tab2show++;
  }

  // Stop the browser following the link
  return false;
}
