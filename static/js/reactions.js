(function () {
    function findFormByType(row, type) {
        var forms = row.querySelectorAll('form');
        for (var i = 0; i < forms.length; i++) {
            var input = forms[i].querySelector('[name="reaction_type"]');
            if (input && input.value === type) return forms[i];
        }
        return null;
    }

    function setupReactionRow(row) {
        row.querySelectorAll('form').forEach(function (form) {
            var typeInput = form.querySelector('[name="reaction_type"]');
            if (!typeInput) return;

            form.addEventListener('submit', function (e) {
                e.preventDefault();

                var type = typeInput.value;
                var likeForm = findFormByType(row, 'like');
                var dislikeForm = findFormByType(row, 'dislike');
                var likeBtn = likeForm && likeForm.querySelector('.reaction-button');
                var dislikeBtn = dislikeForm && dislikeForm.querySelector('.reaction-button');
                var likeCount = likeBtn && likeBtn.querySelector('.reaction-count');
                var dislikeCount = dislikeBtn && dislikeBtn.querySelector('.reaction-count');

                var wasLiked = likeBtn && likeBtn.classList.contains('is-liked');
                var wasDisliked = dislikeBtn && dislikeBtn.classList.contains('is-disliked');

                if (type === 'like' && likeBtn && likeCount) {
                    if (wasLiked) {
                        likeBtn.classList.remove('is-liked');
                        likeCount.textContent = Math.max(0, +likeCount.textContent - 1);
                    } else {
                        likeBtn.classList.add('is-liked');
                        likeCount.textContent = +likeCount.textContent + 1;
                        if (wasDisliked && dislikeBtn && dislikeCount) {
                            dislikeBtn.classList.remove('is-disliked');
                            dislikeCount.textContent = Math.max(0, +dislikeCount.textContent - 1);
                        }
                    }
                } else if (type === 'dislike' && dislikeBtn && dislikeCount) {
                    if (wasDisliked) {
                        dislikeBtn.classList.remove('is-disliked');
                        dislikeCount.textContent = Math.max(0, +dislikeCount.textContent - 1);
                    } else {
                        dislikeBtn.classList.add('is-disliked');
                        dislikeCount.textContent = +dislikeCount.textContent + 1;
                        if (wasLiked && likeBtn && likeCount) {
                            likeBtn.classList.remove('is-liked');
                            likeCount.textContent = Math.max(0, +likeCount.textContent - 1);
                        }
                    }
                }

                fetch(form.action, { method: 'POST', body: new FormData(form) })
                    .catch(function () {});
            });
        });
    }

    document.addEventListener('DOMContentLoaded', function () {
        document.querySelectorAll('.reaction-row').forEach(function (row) {
            setupReactionRow(row);
        });
    });
})();
