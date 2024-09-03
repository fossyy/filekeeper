const socket = new WebSocket('ws://localhost:27000/ws');

socket.onopen = function(event) {
    console.log('WebSocket is open now.');
    socket.send('Hello Server!');
};

socket.onmessage = function(event) {
    try {
        const data = JSON.parse(event.data);
        console.log('Message from server:', data);

        const cpuElement = document.getElementById('cpu_usage');
        if (cpuElement) {
            cpuElement.textContent = `${data.cpu_usage_percent.toFixed(2)}%`;
        }

        const memoryElement = document.getElementById('memory_usage');
        if (memoryElement) {
            memoryElement.textContent = `${data.memory_used_gb.toFixed(2)}/${data.total_memory_gb.toFixed(2)} GB`;
        }

        const uploadElement = document.getElementById('upload_speed');
        if (uploadElement) {
            uploadElement.textContent = `${data.upload_speed_mbps.toFixed(2)}Mbps`;
        }

        const downloadElement = document.getElementById('download_speed');
        if (downloadElement) {
            downloadElement.textContent = `${data.download_speed_mbps.toFixed(2)}Mbps`;
        }
    } catch (error) {
        console.error('Error parsing message data:', error);
    }
};

socket.onerror = function(event) {
    console.error('WebSocket error observed:', event);
};

socket.onclose = function(event) {
    console.log('WebSocket is closed now.');
};