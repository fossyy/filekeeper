if (!window.mySocket) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    window.mySocket = new WebSocket(`${wsProtocol}//${window.location.host}/ws`);

    window.mySocket.onopen = function(event) {
        console.log('WebSocket is open now.');
    };

    window.mySocket.onmessage = async function(event) {
        try {
            const data = JSON.parse(event.data);
            if (data.status === "error") {
                if (data.message === "File Is Different") {
                    ChangeModal("Error", "A file with the same name already exists but has a different hash value. This may indicate that the file is different, despite having the same name. Please verify the file or consider renaming it before proceeding.")
                    toggleModal();
                } else {
                    ChangeModal("Error", "There was an issue with your upload. Please try again later or contact support if the problem persists.")
                    toggleModal();
                }
            } else {
                if (data.action === "UploadNewFile") {
                    if (data.response.Done === false) {
                        const file = window.fileIdMap[data.responseID];
                        addNewUploadElement(file);
                        const fileChunks = await splitFile(file, file.chunkSize);
                        await uploadChunks(file.name, file.size, fileChunks, data.response.Chunk, data.response.ID);
                    } else {
                        ChangeModal("Error", "File already uploaded.")
                        toggleModal();
                    }
                }
            }
        } catch (error) {
            console.error('Error parsing message data:', error);
        }
    };

    window.mySocket.onerror = function(event) {
        console.error('WebSocket error observed:', event);
    };

    window.mySocket.onclose = function(event) {
        console.log('WebSocket is closed now.');
    };
}

function generateUniqueId() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

async function handleFile(file){
    const chunkSize = 2 * 1024 * 1024;
    const chunks = Math.ceil(file.size / chunkSize);
    const fileId = generateUniqueId();
    const startChunk = await file.slice(0, chunkSize).arrayBuffer();
    const endChunk = await file.slice((chunks-1) * chunkSize, file.size).arrayBuffer();
    const startChunkHash = await hash(startChunk)
    const endChunkHash = await  hash(endChunk)
    const data = JSON.stringify({
        "action": "UploadNewFile",
        "name": file.name,
        "size": file.size,
        "chunk": chunks,
        "startHash": startChunkHash,
        "endHash": endChunkHash,
        "requestID": fileId,
    });
    file.chunkSize = chunkSize;
    window.fileIdMap = window.fileIdMap || {};
    window.fileIdMap[fileId] = file;

    window.mySocket.send(data)
}

function addNewUploadElement(file){
    const newDiv = document.createElement('div');
    newDiv.innerHTML = `
      <div class="space-y-4">
          <div class="p-4 flex justify-between items-center">
              <div class="flex items-center space-x-2">
                  <div class="relative">
                      <svg class="w-12 h-12" viewBox="0 0 36 36" xmlns="http://www.w3.org/2000/svg">
                          <circle cx="18" cy="18" r="16" fill="none" class="stroke-current text-gray-200" stroke-width="2"></circle>
                          <circle id="progress-${ file.name }-1" cx="18" cy="18" r="16" fill="none" class="stroke-current text-blue-600" stroke-width="2" stroke-dasharray="100" stroke-dashoffset="100" transform="rotate(-90 18 18)"></circle>
                      </svg>
                      <div class="absolute inset-0 flex items-center justify-center">
                          <span id="progress-${ file.name }-2" class="text-xs font-medium">0%</span>
                      </div>
                  </div>
                  <div class="flex flex-col">
                      <span class="text-base font-medium truncate w-48">${ file.name }</span>
                      <div class="flex items-center gap-x-3 whitespace-nowrap">

                    </div>
                    <div id="progress-${ file.name }-3" class="text-sm text-gray-500">Starting...</div>
                  </div>
              </div>
              <button class="text-blue-500 text-base font-medium">Batal</button>
          </div>
      </div>
      `;
    document.getElementById('FileUploadBoxItem').appendChild(newDiv);
    document.getElementById('uploadBox').classList.remove('hidden');
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

async function hash(arrayBuffer) {
    const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray
        .map((bytes) => bytes.toString(16).padStart(2, '0'))
        .join('');
    return hashHex;
}

async function uploadChunks(name, size, chunks, chunkArray, FileID) {
    let byteUploaded = 0
    let progress1 = document.getElementById(`progress-${name}-1`);
    let progress2 = document.getElementById(`progress-${name}-2`);
    let progress3 = document.getElementById(`progress-${name}-3`);
    let isFailed = false
    for (let index = 0; index < chunks.length; index++) {
        const percentComplete = Math.round((index + 1) / chunks.length * 100);
        const chunk = chunks[index];
        if (!(chunkArray["chunk_"+index])) {
            const formData = new FormData();
            formData.append('name', name);
            formData.append('chunk', chunk);
            formData.append('index', index);
            formData.append('done', false);

            progress1.style.strokeDashoffset = 100 - percentComplete;
            progress2.innerText = `${percentComplete}%`;

            const startTime = performance.now();
            try {
                const request = await fetch(`/file/${FileID}`, {
                    method: 'POST',
                    body: formData
                });
                console.log(request.status)
                if (request.status !== 202) {
                    ChangeModal("Error", "There was an issue with your upload. Please try again later or contact support if the problem persists.")
                    toggleModal();
                    isFailed = true
                    break
                }
            } catch (error) {
                ChangeModal("Error", "There was an issue with your upload. Please try again later or contact support if the problem persists.")
                toggleModal();
                isFailed = true
                break
            }

            const endTime = performance.now();
            const totalTime = (endTime - startTime) / 1000;
            const uploadSpeed = chunk.size / totalTime / 1024 / 1024;
            byteUploaded += chunk.size
            progress3.innerText = `Uploading... ${uploadSpeed.toFixed(2)} MB/s`;
        } else {
            progress1.style.strokeDashoffset = 100 - percentComplete;
            progress2.innerText = `${percentComplete}%`;
            progress3.innerText = `Fixing Missing Byte`;
            byteUploaded += chunk.size
        }
    }
    if (isFailed) {
        progress3.innerText = `Upload Failed`;
    } else {
        progress3.innerText = `Done`;
    }
}

function ChangeModal(title, content) {
    const modalContainer = document.getElementById('modalContainer');
    const modalTitle = modalContainer.querySelector('#modal h2');
    const modalContent = modalContainer.querySelector('.prose');

    modalTitle.textContent = title;
    modalContent.innerHTML = content;
}