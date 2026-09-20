(() => {
    const MAX_IMAGE_SIZE = 20 * 1024 * 1024; // mirrors maxImageSize in handlers/posts.go
    const input = document.getElementById("post-image");
    if (!input) return;

    const field = input.closest(".form-field");

    function showError(message) {
        field.classList.add("has-error");
        let error = field.querySelector(".field-error");
        if (!error) {
            error = document.createElement("p");
            error.id = "post-image-error";
            error.className = "field-error";
            error.setAttribute("role", "alert");
            field.appendChild(error);
        }
        error.textContent = message;
        input.setAttribute("aria-invalid", "true");
        input.setAttribute("aria-describedby", "post-image-error");
    }

    function clearError() {
        field.classList.remove("has-error");
        field.querySelector(".field-error")?.remove();
        input.removeAttribute("aria-invalid");
        input.removeAttribute("aria-describedby");
    }

    input.addEventListener("change", () => {
        const file = input.files[0];
        if (!file) {
            clearError();
            return;
        }
        if (file.size > MAX_IMAGE_SIZE) {
            showError("Image is too big. Max size is 20MB.");
            input.value = ""; // drop the oversized file so it can't be submitted
            return;
        }
        clearError();
    });
})();
