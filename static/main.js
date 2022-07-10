// enables copy to clipboard functionality
const identifiers = document.querySelectorAll('.copy-clipboard');
identifiers.forEach(function(id) {
    id.addEventListener('click', function(event) {
          content = event.target.innerHTML;
          navigator.clipboard.writeText(content);
          notification = event.target.parentNode.querySelector('.copied-notification');
          if (notification === null) {
              let span = document.createElement('span');
              span.className = 'copied-notification';
              span.innerHTML = 'copied!';
              event.target.parentNode.appendChild(span);
          }
    });
});

