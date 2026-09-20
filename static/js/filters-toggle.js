const filtersToggle = document.querySelector("#filters-toggle");
const homeLayout = document.querySelector(".home-layout");

if (filtersToggle && homeLayout) {
    filtersToggle.addEventListener("click", () => {
        const filtersAreHidden = homeLayout.classList.toggle("filters-hidden");

        filtersToggle.setAttribute("aria-expanded", String(!filtersAreHidden));
        filtersToggle.setAttribute("aria-label", filtersAreHidden ? "Show filters" : "Hide filters");
    });
}
