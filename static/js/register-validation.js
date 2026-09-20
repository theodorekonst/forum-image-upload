(() => {
    const password = document.getElementById("register-password");
    const confirmation = document.getElementById("register-confirm-password");
    const confirmationStatus = document.getElementById("confirm-password-status");

    if (!password || !confirmation || !confirmationStatus) return;

    const updateConfirmation = () => {
        confirmationStatus.classList.remove("is-valid", "is-invalid");
        if (!confirmation.value) {
            confirmationStatus.textContent = "";
            return;
        }

        const matches = password.value === confirmation.value;
        confirmationStatus.textContent = matches ? "Passwords match." : "Passwords do not match.";
        confirmationStatus.classList.add(matches ? "is-valid" : "is-invalid");
    };

    password.addEventListener("input", () => {
        updateConfirmation();
    });
    confirmation.addEventListener("input", updateConfirmation);
})();
