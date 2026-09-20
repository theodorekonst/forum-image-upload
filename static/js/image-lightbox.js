(() => {
    const dialog = document.getElementById("image-lightbox");
    const trigger = document.getElementById("post-image-trigger");
    if (!dialog || !trigger) return;

    const closeButton = document.getElementById("image-lightbox-close");

    trigger.addEventListener("click", () => {
        document.body.style.overflow = "hidden";
        dialog.showModal();
    });
    closeButton.addEventListener("click", () => dialog.close());

    // Close when clicking the backdrop, not the image itself.
    dialog.addEventListener("click", event => {
        if (event.target === dialog) dialog.close();
    });

    // Restore background scrolling however the dialog gets closed (button, backdrop, or Esc).
    dialog.addEventListener("close", () => {
        document.body.style.overflow = "";
    });
})();
