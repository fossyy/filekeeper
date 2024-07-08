var isMatch = false
var isValid = false
function validatePasswords() {
    var password = document.getElementById('password').value;
    var confirmPassword = document.getElementById('confirmPassword').value;
    var matchGoodPath = document.getElementById('matchGoodPath');
    var matchBadPath = document.getElementById('matchBadPath');
    var lengthGoodPath = document.getElementById('lengthGoodPath');
    var lengthBadPath = document.getElementById('lengthBadPath');
    var matchSvgContainer = document.getElementById('matchSvgContainer');
    var lengthSvgContainer = document.getElementById('lengthSvgContainer');
    var matchStatusText = document.getElementById('matchStatusText');
    var lengthStatusText = document.getElementById('lengthStatusText');

    if (password === confirmPassword && password.length > 0 && confirmPassword.length > 0 && password.length === confirmPassword.length) {
        matchSvgContainer.classList.remove('bg-red-200');
        matchSvgContainer.classList.add('bg-green-200');
        matchStatusText.classList.remove('text-red-700');
        matchStatusText.classList.add('text-green-700');
        matchGoodPath.style.display = 'inline';
        matchBadPath.style.display = 'none';
        matchStatusText.textContent = "Passwords match";
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

    if ( isValid && isMatch) {
        document.getElementById("submit").disabled = false;
    } else {
        document.getElementById("submit").disabled = true;
    }
}

document.getElementById('password').addEventListener('input', validatePasswords);
document.getElementById('confirmPassword').addEventListener('input', validatePasswords);