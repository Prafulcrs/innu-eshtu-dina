// English ↔ Kannada toggle for UI chrome. Project summaries/notes stay
// English for now — translations below need native review before expanding.
(function () {
  var dict = {
    "How many more days?": "ಇನ್ನೂ ಎಷ್ಟು ದಿನ?",
    "Bengaluru's public scoreboard of promised-versus-delivered infrastructure. Live counters, sourced dates, plain arithmetic.":
      "ಭರವಸೆ ಕೊಟ್ಟದ್ದು ಮತ್ತು ಕೊಟ್ಟಿದ್ದರ ನಡುವಿನ ಬೆಂಗಳೂರಿನ ಸಾರ್ವಜನಿಕ ಸ್ಕೋರ್‌ಬೋರ್ಡ್. ದಿನಾಂಕಗಳು, ಮೂಲಗಳು, ಲೆಕ್ಕ ಮಾತ್ರ.",
    "projects & promises tracked": "ಯೋಜನೆಗಳು & ಭರವಸೆಗಳು",
    "combined days past promised deadlines": "ಒಟ್ಟು ಗಡುವು ಮೀರಿದ ದಿನಗಳು",
    "sanctioned cost of tracked projects": "ಒಟ್ಟು ಮಂಜೂರಾದ ವೆಚ್ಚ",
    "days past the first promised deadline": "ಮೊದಲ ಭರವಸೆಯ ಗಡುವು ಮೀರಿದ ದಿನಗಳು",
    "days since the project began": "ಯೋಜನೆ ಆರಂಭವಾಗಿ ಕಳೆದ ದಿನಗಳು",
    "days since the promise was made": "ಭರವಸೆ ನೀಡಿ ಕಳೆದ ದಿನಗಳು",
    "The word given": "ಕೊಟ್ಟ ಮಾತು",
    "Hall of Finally": "ಕೊನೆಗೂ ಆಯ್ತು!",
    "Where the delays live": "ವಿಳಂಬಗಳ ನಕ್ಷೆ",
    "The promise timeline": "ಭರವಸೆಗಳ ಕಾಲರೇಖೆ",
    "What happened": "ಏನಾಯ್ತು",
    "Sources": "ಮೂಲಗಳು",
    "Methodology": "ವಿಧಾನ",
    "Suggest a project": "ಯೋಜನೆ ಸೂಚಿಸಿ",
    "Stalled": "ಸ್ಥಗಿತ",
    "Crawling": "ಆಮೆಗತಿ",
    "Work resumed": "ಮತ್ತೆ ಶುರು",
    "On the clock": "ಗಡಿಯಾರ ಓಡುತ್ತಿದೆ",
    "Done. Finally.": "ಮುಗೀತು. ಕೊನೆಗೂ."
  };

  var sel = "h1, h2, .tagline, .stat-label, .count-label, .badge, .finally-sub, nav a, .tl-label";

  function apply(kn) {
    document.querySelectorAll(sel).forEach(function (el) {
      if (el.children.length > 0 && !el.dataset.en) return; // only leaf-ish text
      if (!el.dataset.en) el.dataset.en = el.textContent.trim();
      var t = dict[el.dataset.en];
      if (t) el.textContent = kn ? t : el.dataset.en;
    });
    var btn = document.getElementById("lang-toggle");
    if (btn) btn.textContent = kn ? "English" : "ಕನ್ನಡ";
  }

  var btn = document.getElementById("lang-toggle");
  if (btn) {
    btn.addEventListener("click", function (e) {
      e.preventDefault();
      var kn = localStorage.getItem("lang") !== "kn";
      localStorage.setItem("lang", kn ? "kn" : "en");
      apply(kn);
    });
  }
  if (localStorage.getItem("lang") === "kn") apply(true);
})();
