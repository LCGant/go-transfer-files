document.addEventListener('DOMContentLoaded', function () {
    const preloader = document.getElementById('preloader');
    const fileInput = document.getElementById('fileInput');
    const fileName = document.querySelector('.file-name');
    const uploadBtn = document.getElementById('uploadBtn');
    const downloadModal = document.getElementById('downloadModal');
    const closeModalBtn = document.getElementById('closeModal');
    const downloadLinkInput = document.getElementById('downloadLink');
    const copyLinkBtn = document.getElementById('copyLinkBtn');
    const countdownTimer = document.getElementById('countdownTimer');
    const progressBar = document.getElementById('progressBar');
    const progressFill = document.getElementById('progressFill');
    const durationSelect = document.getElementById('durationSelect');
  
    let countdownInterval;
  
    window.onload = function() {
      preloader.classList.add('hidden');
    };
  
    fileInput.addEventListener('change', function() {
      if (fileInput.files.length > 0) {
        fileName.textContent = fileInput.files[0].name;
      } else {
        fileName.textContent = 'No file selected';
      }
    });
  
    function getCookie(name) {
      const value = `; ${document.cookie}`;
      const parts = value.split(`; ${name}=`);
      if (parts.length === 2) return parts.pop().split(';').shift();
    }
  
    uploadBtn.addEventListener('click', function() {
      const file = fileInput.files[0];
      const duration = parseInt(durationSelect.value);
  
      if (!file) {
        alert('Please select a file to upload.');
        return;
      }
  
      if (file.size > 100 * 1024 * 1024) {
        alert('The file exceeds the maximum limit of 100MB.');
        return;
      }
  
      const formData = new FormData();
      formData.append('file', file);
      formData.append('availabilityDuration', duration);
  
      uploadBtn.disabled = true;
      uploadBtn.classList.add('disabled');
      uploadBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Uploading...';
      progressBar.classList.remove('hidden');
      progressFill.style.width = '0%';
  
      const xhr = new XMLHttpRequest();
      xhr.open('POST', '/Files/upload', true);
      xhr.withCredentials = true;
  
      const xsrfToken = getCookie('XSRF-TOKEN');
      if (xsrfToken) {
        xhr.setRequestHeader('X-XSRF-TOKEN', xsrfToken);
      }
  
      xhr.upload.onprogress = function(event) {
        if (event.lengthComputable) {
          const percentComplete = (event.loaded / event.total) * 100;
          progressFill.style.width = `${percentComplete}%`;
        }
      };
  
      xhr.onload = function() {
        uploadBtn.disabled = false;
        uploadBtn.classList.remove('disabled');
        uploadBtn.innerHTML = '<i class="fas fa-cloud-upload-alt"></i> Upload File';
        progressBar.classList.add('hidden');
  
        if (xhr.status === 200) {
          const data = JSON.parse(xhr.responseText);
          console.log('Server response:', data);
          if (data.download_link) {
            downloadLinkInput.value = window.location.origin + data.download_link;
            downloadModal.classList.remove('hidden');
            downloadModal.classList.add('visible');
            startCountdown(duration * 60);
          } else {
            alert('An unexpected error occurred during the upload.');
          }
        } else {
          let errorMessage = 'An error occurred while uploading the file.';
          try {
            const errorData = JSON.parse(xhr.responseText);
            if (errorData.error) {
              errorMessage = `Error: ${errorData.error}`;
            }
          } catch (e) {
            console.error('Error parsing JSON response:', e);
          }
          alert(errorMessage);
        }
      };
  
      xhr.onerror = function() {
        alert('A network error occurred while uploading the file.');
        uploadBtn.disabled = false;
        uploadBtn.classList.remove('disabled');
        uploadBtn.innerHTML = '<i class="fas fa-cloud-upload-alt"></i> Upload File';
        progressBar.classList.add('hidden');
      };
  
      xhr.send(formData);
    });
  
    copyLinkBtn.addEventListener('click', function() {
      downloadLinkInput.select();
      downloadLinkInput.setSelectionRange(0, 99999);
      navigator.clipboard.writeText(downloadLinkInput.value)
        .then(() => {
          alert('Download link copied to clipboard!');
        })
        .catch(() => {
          alert('Failed to copy the link.');
        });
    });
  
    closeModalBtn.addEventListener('click', function() {
      downloadModal.classList.add('hidden');
      downloadModal.classList.remove('visible');
    });
  
    const uploadBox = document.querySelector('.upload-box');
    uploadBox.addEventListener('dragover', function(e) {
      e.preventDefault();
      e.stopPropagation();
      uploadBox.classList.add('dragover');
    });
    uploadBox.addEventListener('dragleave', function(e) {
      e.preventDefault();
      e.stopPropagation();
      uploadBox.classList.remove('dragover');
    });
    uploadBox.addEventListener('drop', function(e) {
      e.preventDefault();
      e.stopPropagation();
      uploadBox.classList.remove('dragover');
      const files = e.dataTransfer.files;
      if (files.length > 0) {
        fileInput.files = files;
        fileName.textContent = files[0].name;
      }
    });
  
    function startCountdown(duration) {
      let timer = duration;
      updateCountdown(timer);
      countdownInterval = setInterval(function () {
        timer--;
        if (timer < 0) {
          clearInterval(countdownInterval);
          countdownTimer.textContent = 'The link has expired.';
          downloadLinkInput.value = 'Link expired';
          downloadLinkInput.disabled = true;
          copyLinkBtn.disabled = true;
          setTimeout(() => {
            downloadModal.classList.add('hidden');
            downloadModal.classList.remove('visible');
          }, 2000);
        } else {
          updateCountdown(timer);
        }
      }, 1000);
    }
  
    function updateCountdown(timer) {
      countdownTimer.textContent = `Time Remaining: ${formatTime(timer)}`;
    }
  
    function formatTime(seconds) {
      const minutes = String(Math.floor(seconds / 60)).padStart(2, '0');
      const secs = String(seconds % 60).padStart(2, '0');
      return `${minutes}:${secs}`;
    }
  });
  