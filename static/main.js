// enables copy to clipboard functionality
const identifiers = document.querySelectorAll('.copy-clipboard');
identifiers.forEach(function(id) {
    id.addEventListener('click', function(event) {
        // find related value to copy
        group = event.target.closest('.group');
        val = group.querySelector('.value').textContent;
        // copy value to clipboard
        navigator.clipboard.writeText(val);
        // trigger brief notification
        orig = event.target.innerHTML;
        event.target.innerHTML = '[copied!]';
        setTimeout(function() {
            // revert original content
            event.target.innerHTML = orig;
        }, 1000);
    });
});

