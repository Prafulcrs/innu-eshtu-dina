(function () {
  var IST = "T00:00:00+05:30";

  function fmt(n) {
    return n.toLocaleString("en-IN");
  }

  function daysSince(dateStr) {
    var from = new Date(dateStr + IST).getTime();
    return Math.max(0, Math.floor((Date.now() - from) / 86400000));
  }

  // static day counters (index cards)
  document.querySelectorAll("[data-days-from]").forEach(function (el) {
    el.textContent = fmt(daysSince(el.getAttribute("data-days-from")));
  });

  // live tickers (project hero): "N days HH:MM:SS"
  document.querySelectorAll("[data-ticker-from]").forEach(function (el) {
    var from = new Date(el.getAttribute("data-ticker-from") + IST).getTime();
    function pad(n) { return String(n).padStart(2, "0"); }
    function tick() {
      var s = Math.max(0, Math.floor((Date.now() - from) / 1000));
      var d = Math.floor(s / 86400); s -= d * 86400;
      var h = Math.floor(s / 3600); s -= h * 3600;
      var m = Math.floor(s / 60); s -= m * 60;
      el.textContent = fmt(d) + " days " + pad(h) + ":" + pad(m) + ":" + pad(s);
    }
    tick();
    setInterval(tick, 1000);
  });
})();
