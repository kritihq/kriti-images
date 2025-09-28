function showTab(evt, tabId) {
  // Hide all tab contents in the same section
  const section = evt.target.closest(".transformation-section");
  const tabContents = section.querySelectorAll(".tab-content");
  const tabs = section.querySelectorAll(".tab");

  // Remove active class from all tabs and contents
  tabContents.forEach((content) => content.classList.remove("active"));
  tabs.forEach((tab) => tab.classList.remove("active"));

  // Show selected tab
  document.getElementById(tabId).classList.add("active");
  evt.target.classList.add("active");
}
