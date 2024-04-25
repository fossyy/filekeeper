var isMatch = false
var isValid = false
var isSecure = false
function validatePasswords() {
    var password = document.getElementById('password').value;
    var confirmPassword = document.getElementById('confirmPassword').value;
    var matchGoodPath = document.getElementById('matchGoodPath');
    var matchBadPath = document.getElementById('matchBadPath');
    var lengthGoodPath = document.getElementById('lengthGoodPath');
    var lengthBadPath = document.getElementById('lengthBadPath');
    var requirementsGoodPath = document.getElementById('requirementsGoodPath');
    var requirementsBadPath = document.getElementById('requirementsBadPath');
    var matchSvgContainer = document.getElementById('matchSvgContainer');
    var lengthSvgContainer = document.getElementById('lengthSvgContainer');
    var requirementsSvgContainer = document.getElementById('requirementsSvgContainer');
    var matchStatusText = document.getElementById('matchStatusText');
    var lengthStatusText = document.getElementById('lengthStatusText');
    var requirementsStatusText = document.getElementById('requirementsStatusText');
    var symbolRegex = /[!@#$%^&*]/;
    var uppercaseRegex = /[A-Z]/;
    var numberRegex = /\d/g;


    if (password === confirmPassword && password.length > 0 && confirmPassword.length > 0 && password.length === confirmPassword.length) {
        matchSvgContainer.classList.remove('bg-red-200');
        matchSvgContainer.classList.add('bg-green-200');
        matchStatusText.classList.remove('text-red-700');
        matchStatusText.classList.add('text-green-700');
        matchGoodPath.style.display = 'inline';
        matchBadPath.style.display = 'none';
        matchStatusText.textContent = "Passwords match";
        console.log("anjay")
        isMatch = true
    } else {
        matchSvgContainer.classList.remove('bg-green-200');
        matchSvgContainer.classList.add('bg-red-200');
        matchStatusText.classList.remove('text-green-700');
        matchStatusText.classList.add('text-red-700');
        matchGoodPath.style.display = 'none';
        matchBadPath.style.display = 'inline';
        matchStatusText.textContent = "Passwords do not match";
        isMatch = false
    }

    if (password.length >= 8) {
        lengthSvgContainer.classList.remove('bg-red-200');
        lengthSvgContainer.classList.add('bg-green-200');
        lengthStatusText.classList.remove('text-red-700');
        lengthStatusText.classList.add('text-green-700');
        lengthGoodPath.style.display = 'inline';
        lengthBadPath.style.display = 'none';
        lengthStatusText.textContent = "Password length meets requirement";
        isValid = true
    } else {
        lengthSvgContainer.classList.remove('bg-green-200');
        lengthSvgContainer.classList.add('bg-red-200');
        lengthStatusText.classList.remove('text-green-700');
        lengthStatusText.classList.add('text-red-700');
        lengthGoodPath.style.display = 'none';
        lengthBadPath.style.display = 'inline';
        lengthStatusText.textContent = "Password length must be at least 8 characters";
        isValid = false
    }

    var symbolCheck = symbolRegex.test(password);
    var uppercaseCheck = uppercaseRegex.test(password);
    var numberCount = (password.match(numberRegex) || []).length;

    if (symbolCheck && uppercaseCheck && numberCount >= 3) {
        requirementsSvgContainer.classList.remove('bg-red-200');
        requirementsSvgContainer.classList.add('bg-green-200');
        requirementsStatusText.classList.remove('text-red-700');
        requirementsStatusText.classList.add('text-green-700');
        requirementsGoodPath.style.display = 'inline';
        requirementsBadPath.style.display = 'none';
        requirementsStatusText.textContent = "Password meets additional requirements";
        isSecure = true
    } else {
        requirementsSvgContainer.classList.remove('bg-green-200');
        requirementsSvgContainer.classList.add('bg-red-200');
        requirementsStatusText.classList.remove('text-green-700');
        requirementsStatusText.classList.add('text-red-700');
        requirementsGoodPath.style.display = 'none';
        requirementsBadPath.style.display = 'inline';
        requirementsStatusText.textContent = "The password must contain at least one symbol (!@#$%^&*), one uppercase letter, and three numbers";
        isSecure = false
    }

    if (isSecure && isValid && isSecure && password === confirmPassword) {
        document.getElementById("submit").disabled = false;
    } else {
        document.getElementById("submit").disabled = true;
    }
}

document.getElementById('password').addEventListener('input', validatePasswords);
document.getElementById('confirmPassword').addEventListener('input', validatePasswords);