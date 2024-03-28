document.addEventListener("dragover", function (event) {
    event.preventDefault();
});

document.addEventListener("drop", async function (event) {
    event.preventDefault();
    const file = event.dataTransfer.files[0]
    const chunkSize = 2 * 1024 * 1024;
    const chunks = Math.ceil(file.size / chunkSize);
    const data = JSON.stringify({
        "name": file.name,
        "size": file.size,
        "chunk": chunks,
    });

    fetch('http://localhost:8000/upload/init', {
        method: 'POST',
        body: data,
    }).then(async response => {
        console.log('File uploaded successfully.');
        const fileChunks = await splitFile(file, chunkSize);
        await uploadChunks(file.name ,fileChunks);
    }).catch(error => {
        console.error('Error uploading file:', error);
    });

});

async function splitFile(file, chunkSize) {
    const fileSize = file.size;
    const chunks = Math.ceil(fileSize / chunkSize);
    const fileChunks = [];

    for (let i = 0; i < chunks; i++) {
        const start = i * chunkSize;
        const end = Math.min(fileSize, start + chunkSize);
        const chunk = file.slice(start, end);
        fileChunks.push(chunk);
    }

    return fileChunks;
}

async function uploadChunks(name, chunks) {
    const uploadPromises = chunks.map((chunk, index) => {
        const formData = new FormData();
        formData.append('name', name);
        formData.append('chunk', chunk);
        formData.append('index', index);
        formData.append('done', false);
        const percentComplete = Math.round(Math.round(index + 1) / chunks.length * 100);
        var progress1 = document.getElementById("progres1");
        var progress2 = document.getElementById("progres2");
        var progress3 = document.getElementById("progres3");
        progress1.setAttribute("aria-valuenow", percentComplete);
        progress2.style.width = `${percentComplete}%`;
        progress3.innerText = `${percentComplete}%`;
        console.log(percentComplete);
        return fetch('http://localhost:8000/upload', {
            method: 'POST',
            body: formData
        });
    });

    await Promise.all(uploadPromises);
    const formData = new FormData();
    formData.append('name', name);
    formData.append('done', true);
    return fetch('http://localhost:8000/upload', {
        method: 'POST',
        body: formData
    });
}


