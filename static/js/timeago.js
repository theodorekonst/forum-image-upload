(function () {
    function timeAgo(date) {
        var seconds = Math.floor((new Date() - date) / 1000);
        if (seconds < 45) return 'just now';
        if (seconds < 90) return '1 minute ago';
        var minutes = Math.round(seconds / 60);
        if (minutes < 60) return minutes + ' minutes ago';
        var hours = Math.round(minutes / 60);
        if (hours < 24) return hours === 1 ? '1 hour ago' : hours + ' hours ago';
        var days = Math.round(hours / 24);
        if (days === 1) return 'yesterday';
        if (days < 7) return days + ' days ago';
        var weeks = Math.round(days / 7);
        if (weeks < 5) return weeks === 1 ? '1 week ago' : weeks + ' weeks ago';
        var months = Math.round(days / 30);
        if (months < 12) return months === 1 ? '1 month ago' : months + ' months ago';
        var years = Math.round(days / 365);
        return years === 1 ? '1 year ago' : years + ' years ago';
    }

    document.addEventListener('DOMContentLoaded', function () {
        document.querySelectorAll('time[datetime]').forEach(function (el) {
            var date = new Date(el.getAttribute('datetime'));
            if (!isNaN(date.getTime())) {
                el.textContent = timeAgo(date);
            }
        });
    });
})();
