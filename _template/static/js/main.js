document.body.addEventListener("theme-updated", function () {
  var select = document.querySelector("[name='theme']");
  if (select) {
    document.documentElement.dataset.theme = select.value;
  }
});

document.body.addEventListener("htmx:beforeSwap", function (evt) {
  if (evt.detail.xhr.status >= 400) {
    evt.detail.shouldSwap = true;
    evt.detail.isError = false;
  }
});
