document.addEventListener('DOMContentLoaded', () => {
    const dropArea = document.getElementById('drop-area');
    const fileElem = document.getElementById('fileElem');
    const filterInput = document.getElementById('filterInput');
    const logList = document.getElementById('log-list');
    const detailView = document.querySelector('#detail-view pre code');

    let allLogs = [];
    let selectedIndex = 0;
    let ws;

    function connectWebSocket() {
        ws = new WebSocket(`ws://${window.location.host}/ws`);

        ws.onopen = () => {
            console.log('Connected to LogLens server');
        };

        ws.onmessage = (event) => {
            const log = JSON.parse(event.data);
            allLogs.push(log);
            renderLogList();
        };

        ws.onclose = () => {
            console.log('Disconnected. Attempting to reconnect...');
            setTimeout(connectWebSocket, 3000); // Reconnect after 3 seconds
        };
    }
    
    // --- Event Listeners ---
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, () => dropArea.classList.add('highlight'), false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, () => dropArea.classList.remove('highlight'), false);
    });

    dropArea.addEventListener('drop', handleDrop, false);
    fileElem.addEventListener('change', (e) => handleFiles(e.target.files));
    filterInput.addEventListener('keyup', renderLogList);

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    function handleDrop(e) {
        let dt = e.dataTransfer;
        let files = dt.files;
        handleFiles(files);
    }

    function handleFiles(files) {
        if (files.length === 0) return;
        const file = files[0];
        const formData = new FormData();
        formData.append('logfile', file);

        // Reset state for new file
        allLogs = [];
        selectedIndex = 0;
        renderLogList();
        detailView.textContent = "Processing log file...";

        fetch('/upload', {
            method: 'POST',
            body: formData,
        })
        .then(response => response.json())
        .then(data => {
            console.log('Upload complete:', data);
            detailView.textContent = `${data.lines_processed} lines processed. Waiting for logs...`;
        })
        .catch(error => {
            console.error('Error:', error);
            detailView.textContent = `Error processing file: ${error}`;
        });
    }

    function renderLogList() {
        const filterText = filterInput.value.toLowerCase();
        const filteredLogs = allLogs.filter(log => {
            if (!filterText) return true;
            return JSON.stringify(log).toLowerCase().includes(filterText);
        });

        logList.innerHTML = ''; // Clear list

        filteredLogs.forEach((log, index) => {
            const item = document.createElement('div');
            item.className = 'log-item';
            
            const level = log.level || 'unknown';
            const message = log.message || 'No message';
            item.textContent = `[${level.toUpperCase()}] ${message}`;

            if (index === selectedIndex) {
                item.classList.add('selected');
                renderDetailView(log);
            }

            item.addEventListener('click', () => {
                selectedIndex = index;
                renderLogList(); // Re-render to update selection
            });

            logList.appendChild(item);
        });

        if (filteredLogs.length === 0) {
            detailView.textContent = "No logs match the filter.";
        }
    }

    function renderDetailView(log) {
        detailView.textContent = JSON.stringify(log, null, 2);
    }
    
    // Initial connection
    connectWebSocket();
});