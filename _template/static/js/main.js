document.body.addEventListener("theme-updated", function () {
  var select = document.querySelector("[name='theme']");
  if (select) {
    document.documentElement.dataset.theme = select.value;
  }
});
