document.addEventListener('DOMContentLoaded', function () {
    const openDownloadModalBtn = document.getElementById('openDownloadModal');
    const downloadModal = document.getElementById('downloadModal');
    const closeDownloadModalBtn = document.getElementById('closeDownloadModal');
    const userDownloadLinkInput = document.getElementById('userDownloadLink');
    const downloadBtn = document.getElementById('downloadBtn');

    function showDownloadModal() {
        downloadModal.classList.remove('hidden');
        downloadModal.classList.add('visible');
    }

    function hideDownloadModal() {
        downloadModal.classList.remove('visible');
        downloadModal.classList.add('hidden');
        userDownloadLinkInput.value = '';
    }

    openDownloadModalBtn.addEventListener('click', function () {
        showDownloadModal();
    });

    closeDownloadModalBtn.addEventListener('click', function () {
        hideDownloadModal();
    });

    window.addEventListener('click', function (event) {
        if (event.target === downloadModal) {
            hideDownloadModal();
        }
    });

    downloadBtn.addEventListener('click', function () {
        const userLink = userDownloadLinkInput.value.trim();

        if (!userLink) {
            showNotification('Please enter a download link.', 3000);
            return;
        }

        try {
            const url = new URL(userLink);
            const token = url.searchParams.get('token');

            if (!token) {
                showNotification('Invalid link. The download token is missing.', 3000);
                return;
            }

            window.location.href = `/Files/download?token=${encodeURIComponent(token)}`;

            hideDownloadModal();
        } catch (error) {
            showNotification('Invalid URL. Please provide a valid download link.', 3000);
        }
    });

    function showNotification(message, timeout = 3000) {
        const notificationMessage = document.getElementById('notificationMessage');
        const notificationBalloon = document.getElementById('notificationBalloon');
        const notificationIcon = document.getElementById('notificationIcon');

        if (notificationMessage) {
            notificationMessage.innerText = message;
        }

        if (notificationBalloon && notificationIcon) {
            notificationBalloon.style.display = 'flex';
            notificationBalloon.style.flexDirection = 'column';
            notificationBalloon.style.gap = '10px';

            notificationIcon.style.display = 'flex';
            notificationIcon.style.flexDirection = 'column';
            notificationIcon.style.gap = '10px';

            const balloonHeight = notificationBalloon.offsetHeight;
            notificationIcon.style.top = `${20 + balloonHeight + 10}px`;

            setTimeout(closeNotification, timeout);
        }
    }

    function closeNotification() {
        const notificationBalloon = document.getElementById('notificationBalloon');
        const notificationIcon = document.getElementById('notificationIcon');
        if (notificationBalloon) notificationBalloon.style.display = 'none';
        if (notificationIcon) notificationIcon.style.display = 'none';
    }
});
