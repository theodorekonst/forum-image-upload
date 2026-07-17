document.querySelectorAll(".form-field").forEach(field => {
    const control = field.querySelector("input, textarea");
    if (!control) return;

    control.addEventListener("input", () => {
        const error = field.querySelector(".field-error");
        if (!error) return;

        field.classList.remove("has-error");
        control.removeAttribute("aria-invalid");

        const describedBy = (control.getAttribute("aria-describedby") || "")
            .split(/\s+/)
            .filter(id => id && id !== error.id);

        if (describedBy.length) {
            control.setAttribute("aria-describedby", describedBy.join(" "));
        } else {
            control.removeAttribute("aria-describedby");
        }

        error.remove();
    }, { once: true });
});

document.querySelectorAll(".content-form").forEach(form => {
    const error = form.querySelector(":scope > .form-error");
    if (!error) return;

    form.addEventListener("input", () => error.remove(), { once: true });
});
