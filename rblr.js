"use strict";

const myStackItem = "odoStack";

function loadPage(x) {

  window.location.href = x;
  
}

// Called during Odo capture when input is entered
function oi(obj) {
  obj.classList.remove("oc");
  obj.classList.add("oi");

  // autosave handler
  if (obj.timer) {
    clearTimeout(obj.timer);
  }
  obj.timer = setTimeout(saveOdo,3000,obj);
}

// Called during Odo capture when input is complete
function oc(obj) {
  saveOdo(obj);
}

function saveOdo(obj) {

  if (obj.timer) {
    clearTimeout(obj.timer);
  }

  let ent = obj.getAttribute("data-e");
  let url = "/setodo?e=" + ent + "&f=" + obj.name + "&v=" + obj.value;

  let newTrans = {};
  newTrans.url = url;
  newTrans.obj = obj.id;
  newTrans.sent = false;

  const stackx = sessionStorage.getItem(myStackItem);
  let stack = [];
  if (stackx != null) {
    stack = JSON.parse(stackx);
  }
  stack.push(newTrans);
  sessionStorage.setItem(myStackItem, JSON.stringify(stack));
  obj.classList.remove("oi");
  obj.classList.add("oc");

}
// Called periodically to send outstanding updates to backend server
function sendTransactions() {
  let stackx = sessionStorage.getItem(myStackItem);
  if (stackx == null) return;

  let stack = JSON.parse(stackx);

  let N = stack.length;

  if (N < 1) return;

  for (let i = 0; i < N; i++) {
    if (stack[i].sent) continue;

    console.log(stack[i].url);

    let xhttp = new XMLHttpRequest();
    xhttp.onerror = function () {
      return; // Probably means the network's not available
    };
    xhttp.onload = function () {
      let ok = new RegExp("\\W*ok\\W*");
      if (xhttp.status == 200) {
        console.log("{" + this.responseText.substring(0, 30) + "}");
        if (!ok.test(this.responseText)) {
          console.log("UPDATE_FAILED");
          return;
        } else {
          stack[i].sent = true;
          sessionStorage.setItem(myStackItem, JSON.stringify(stack));
          document.getElementById(stack[i].obj).classList.replace("oc", "ok");
        }
      }
    };
    xhttp.open("GET", stack[i].url, true);
    xhttp.send();
  }
}
