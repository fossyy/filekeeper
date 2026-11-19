var isMatch = false;
var isValid = false;
var hasUppercase = false;
var hasNumber = false;

function validatePasswords() {
    var checkContainer = document.getElementById('checkContainer')
    var passwordInput = document.getElementById('password');
    var confirmPasswordInput = document.getElementById('confirmPassword');
    var matchGoodPath = document.getElementById('matchGoodPath');
    var matchBadPath = document.getElementById('matchBadPath');
    var lengthGoodPath = document.getElementById('lengthGoodPath');
    var lengthBadPath = document.getElementById('lengthBadPath');
    var uppercaseGoodPath = document.getElementById('uppercaseGoodPath');
    var uppercaseBadPath = document.getElementById('uppercaseBadPath');
    var matchSvgContainer = document.getElementById('matchSvgContainer');
    var lengthSvgContainer = document.getElementById('lengthSvgContainer');
    var uppercaseSvgContainer = document.getElementById('uppercaseSvgContainer');
    var matchStatusText = document.getElementById('matchStatusText');
    var lengthStatusText = document.getElementById('lengthStatusText');
    var uppercaseStatusText = document.getElementById('uppercaseStatusText');
    var numberSvgContainer = document.getElementById('numberSvgContainer');
    var numberGoodPath = document.getElementById('numberGoodPath');
    var numberBadPath = document.getElementById('numberBadPath');
    var numberStatusText = document.getElementById('numberStatusText');
    var submitButton = document.getElementById('submit');
    if (!checkContainer || !passwordInput || !confirmPasswordInput ||
        !matchGoodPath || !matchBadPath || !lengthGoodPath || !lengthBadPath ||
        !uppercaseGoodPath || !uppercaseBadPath || !matchSvgContainer ||
        !lengthSvgContainer || !uppercaseSvgContainer || !matchStatusText ||
        !lengthStatusText || !uppercaseStatusText || !numberSvgContainer ||
        !numberGoodPath || !numberBadPath || !numberStatusText || !submitButton) {
        return;
    }
    var password = passwordInput.value;
    var confirmPassword = confirmPasswordInput.value;
    if (password.length > 0) {
        checkContainer.classList.remove("hidden")
    } else {
        checkContainer.classList.add("hidden")
    }
    if (password === confirmPassword && password.length > 0 && confirmPassword.length > 0 && password.length === confirmPassword.length) {
        matchSvgContainer.classList.remove('bg-red-200');
        matchSvgContainer.classList.add('bg-green-200');
        matchStatusText.classList.remove('text-red-700');
        matchStatusText.classList.add('text-green-700');
        matchGoodPath.style.display = 'inline';
        matchBadPath.style.display = 'none';
        matchStatusText.textContent = "Passwords match";
        isMatch = true;
    } else {
        matchSvgContainer.classList.remove('bg-green-200');
        matchSvgContainer.classList.add('bg-red-200');
        matchStatusText.classList.remove('text-green-700');
        matchStatusText.classList.add('text-red-700');
        matchGoodPath.style.display = 'none';
        matchBadPath.style.display = 'inline';
        matchStatusText.textContent = "Passwords do not match";
        isMatch = false;
    }

    if (password.length >= 8) {
        lengthSvgContainer.classList.remove('bg-red-200');
        lengthSvgContainer.classList.add('bg-green-200');
        lengthStatusText.classList.remove('text-red-700');
        lengthStatusText.classList.add('text-green-700');
        lengthGoodPath.style.display = 'inline';
        lengthBadPath.style.display = 'none';
        lengthStatusText.textContent = "Password length meets requirement";
        isValid = true;
    } else {
        lengthSvgContainer.classList.remove('bg-green-200');
        lengthSvgContainer.classList.add('bg-red-200');
        lengthStatusText.classList.remove('text-green-700');
        lengthStatusText.classList.add('text-red-700');
        lengthGoodPath.style.display = 'none';
        lengthBadPath.style.display = 'inline';
        lengthStatusText.textContent = "Password length must be at least 8 characters";
        isValid = false;
    }

    if (/[A-Z]/.test(password)) {
        uppercaseSvgContainer.classList.remove('bg-red-200');
        uppercaseSvgContainer.classList.add('bg-green-200');
        uppercaseStatusText.classList.remove('text-red-700');
        uppercaseStatusText.classList.add('text-green-700');
        uppercaseGoodPath.style.display = 'inline';
        uppercaseBadPath.style.display = 'none';
        uppercaseStatusText.textContent = "Password contains an uppercase letter";
        hasUppercase = true;
    } else {
        uppercaseSvgContainer.classList.remove('bg-green-200');
        uppercaseSvgContainer.classList.add('bg-red-200');
        uppercaseStatusText.classList.remove('text-green-700');
        uppercaseStatusText.classList.add('text-red-700');
        uppercaseGoodPath.style.display = 'none';
        uppercaseBadPath.style.display = 'inline';
        uppercaseStatusText.textContent = "Password must contain at least one uppercase letter";
        hasUppercase = false;
    }

    if (/[0-9]/.test(password)) {
        numberSvgContainer.classList.remove('bg-red-200');
        numberSvgContainer.classList.add('bg-green-200');
        numberStatusText.classList.remove('text-red-700');
        numberStatusText.classList.add('text-green-700');
        numberGoodPath.style.display = 'inline';
        numberBadPath.style.display = 'none';
        numberStatusText.textContent = 'Password contains a number';
        hasNumber = true;
    } else {
        numberSvgContainer.classList.remove('bg-green-200');
        numberSvgContainer.classList.add('bg-red-200');
        numberStatusText.classList.remove('text-green-700');
        numberStatusText.classList.add('text-red-700');
        numberGoodPath.style.display = 'none';
        numberBadPath.style.display = 'inline';
        numberStatusText.textContent = 'Password must contain a number';
        hasNumber = false;
    }

    if (isValid && isMatch && hasUppercase && hasNumber) {
        submitButton.disabled = false;
    } else {
        submitButton.disabled = true;
    }
}

const passwordInput = document.getElementById('password');
const confirmationInput = document.getElementById('confirmPassword');
if (passwordInput && confirmationInput) {
    passwordInput.addEventListener('input', validatePasswords);
    confirmationInput.addEventListener('input', validatePasswords);
}
