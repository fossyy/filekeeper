document.addEventListener("dragover", function (event) {
    event.preventDefault();
});

document.addEventListener("drop", async function (event) {
    event.preventDefault();
    const file = event.dataTransfer.files[0]
    await handleFile(file)
});

document.getElementById('dropzone-file').addEventListener('change', async function(event) {
    event.preventDefault();
    const file = event.target.files[0]
    await handleFile(file)
});

async function handleFile(file){
    const chunkSize = 2 * 1024 * 1024;
    const chunks = Math.ceil(file.size / chunkSize);
    const data = JSON.stringify({
        "name": file.name,
        "size": file.size,
        "chunk": chunks,
    });

    fetch('/upload/init', {
        method: 'POST',
        body: data,
    }).then(async response => {
        console.log(response.status);
        console.log('File uploaded successfully.');
        if (response.status !== 200) {
            alert("file already exist")
            return
        }
        addNewUploadElement(file)
        const fileChunks = await splitFile(file, chunkSize);
        await uploadChunks(file.name ,fileChunks);
    }).catch(error => {
        console.error('Error uploading file:', error);
    });
}

function addNewUploadElement(file){
    const newDiv = document.createElement('div');
    newDiv.innerHTML = `
    <div class="mb-2 flex justify-between items-center">
      <div class="flex items-center gap-x-3">
        <svg xmlns="http://www.w3.org/2000/svg" x="0px" y="0px" width="100" height="100" viewBox="0 0 48 48">
<path fill="#90CAF9" d="M40 45L8 45 8 3 30 3 40 13z"></path><path fill="#E1F5FE" d="M38.5 14L29 14 29 4.5z"></path>
</svg>
        <div>
          <p class="text-sm font-medium text-gray-800 dark:text-white">${file.name}</p>
          <p class="text-xs text-gray-500 dark:text-gray-500">${convertFileSize(file.size)}</p>
        </div>
      </div>
      <div class="inline-flex items-center gap-x-2">
        <a class="text-gray-500 hover:text-gray-800" href="#">
          <svg class="flex-shrink-0 size-4" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect width="4" height="16" x="6" y="4" />
            <rect width="4" height="16" x="14" y="4" />
          </svg>
        </a>
        <a class="text-gray-500 hover:text-gray-800" href="#">
          <svg class="flex-shrink-0 size-4" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 6h18" />
            <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
            <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
            <line x1="10" x2="10" y1="11" y2="17" />
            <line x1="14" x2="14" y1="11" y2="17" />
          </svg>
        </a>
      </div>
    </div>
    
    <div class="flex items-center gap-x-3 whitespace-nowrap">
      <div id="progress-${file.name}-1" class="flex w-full h-2 bg-gray-200 rounded-full overflow-hidden dark:bg-gray-700"
        role="progressbar" aria-valuenow="100" aria-valuemin="0" aria-valuemax="100">
        <div id="progress-${file.name}-2"
          class="flex flex-col justify-center rounded-full overflow-hidden bg-teal-500 text-xs text-white text-center whitespace-nowrap transition duration-500">
        </div>
      </div>
      <div class="w-6 text-end">
        <span id="progress-${file.name}-3" class="text-sm text-gray-800 dark:text-white">Starting...</span>
      </div>
    </div>
      `;
    document.getElementById('container').appendChild(newDiv);
}

function convertFileSize(sizeInBytes) {
    if (sizeInBytes < 1024) {
        return sizeInBytes + ' B';
    } else if (sizeInBytes < 1024 * 1024) {
        return (sizeInBytes / 1024).toFixed(2) + ' KB';
    } else if (sizeInBytes < 1024 * 1024 * 1024) {
        return (sizeInBytes / (1024 * 1024)).toFixed(2) + ' MB';
    } else {
        return (sizeInBytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    }
}

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
    for (let index = 0; index < chunks.length; index++) {
        const chunk = chunks[index];
        const formData = new FormData();
        formData.append('name', name);
        formData.append('chunk', chunk);
        formData.append('index', index);
        formData.append('done', false);
        const percentComplete = Math.round((index + 1) / chunks.length * 100);
        var progress1 = document.getElementById(`progress-${name}-1`);
        var progress2 = document.getElementById(`progress-${name}-2`);
        var progress3 = document.getElementById(`progress-${name}-3`);
        progress1.setAttribute("aria-valuenow", percentComplete);
        progress2.style.width = `${percentComplete}%`;
        progress3.innerText = `${percentComplete}%`;
        await fetch('/upload', {
            method: 'POST',
            body: formData
        });

    }

    const formData = new FormData();
    formData.append('name', name);
    formData.append('done', true);
    return fetch('/upload', {
        method: 'POST',
        body: formData
    });
}


