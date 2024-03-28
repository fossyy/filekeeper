function copyToClipboard(tdId) {
    var tdContent = document.getElementById(tdId).innerText;

    var tempTextArea = document.createElement("textarea");
    tempTextArea.value = tdContent;

    document.body.appendChild(tempTextArea);

    tempTextArea.select();
    tempTextArea.setSelectionRange(0, 99999);

    document.execCommand("copy");

    document.body.removeChild(tempTextArea);

    alert("Copied the text from " + tdId + ": " + tdContent);
}