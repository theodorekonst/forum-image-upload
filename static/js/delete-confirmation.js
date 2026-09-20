(() => {
    const dialog = document.getElementById("delete-confirmation");
    if (!dialog) return;

    const title = document.getElementById("delete-confirmation-title");
    const message = document.getElementById("delete-confirmation-message");
    const cancelButton = dialog.querySelector(".dialog-cancel-button");
    const deleteButton = dialog.querySelector(".dialog-delete-button");
    let pendingForm = null;

    document.querySelectorAll(".delete-form").forEach(form => {
        form.addEventListener("submit", event => {
            if (form.dataset.confirmed === "true") return;

            event.preventDefault();
            pendingForm = form;
            title.textContent = form.dataset.confirmTitle || "Delete?";
            message.textContent = form.dataset.confirmMessage || "This action cannot be undone.";
            dialog.showModal();
            cancelButton.focus();
        });
    });

    cancelButton.addEventListener("click", () => dialog.close());

    deleteButton.addEventListener("click", () => {
        if (!pendingForm) return;
        pendingForm.dataset.confirmed = "true";
        dialog.close();
        pendingForm.requestSubmit();
    });

    dialog.addEventListener("close", () => {
        if (pendingForm?.dataset.confirmed !== "true") pendingForm = null;
    });
})();
