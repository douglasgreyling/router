// Generate table of contents from headings
document.addEventListener('DOMContentLoaded', function() {
  const toc = document.getElementById('toc');
  if (!toc) return;

  const article = document.querySelector('article');
  if (!article) return;

  const headings = article.querySelectorAll('h2, h3');
  if (headings.length === 0) return;

  headings.forEach((heading, index) => {
    // Add ID if not present
    if (!heading.id) {
      heading.id = heading.textContent
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/(^-|-$)/g, '');
    }

    // Create TOC item
    const li = document.createElement('li');
    const a = document.createElement('a');

    if (heading.tagName === 'H3') {
      li.classList.add('level-3');
    }

    a.href = '#' + heading.id;
    a.textContent = heading.textContent;
    li.appendChild(a);
    toc.appendChild(li);
  });

  // Highlight active section on scroll
  const tocLinks = toc.querySelectorAll('a');

  function highlightTOC() {
    let current = '';

    headings.forEach(heading => {
      const sectionTop = heading.offsetTop;
      if (window.scrollY >= sectionTop - 100) {
        current = heading.id;
      }
    });

    tocLinks.forEach(link => {
      link.classList.remove('active');
      if (link.getAttribute('href') === '#' + current) {
        link.classList.add('active');
      }
    });
  }

  window.addEventListener('scroll', highlightTOC);
  highlightTOC();

  // Smooth scroll
  tocLinks.forEach(link => {
    link.addEventListener('click', function(e) {
      e.preventDefault();
      const targetId = this.getAttribute('href').substring(1);
      const targetElement = document.getElementById(targetId);
      if (targetElement) {
        window.scrollTo({
          top: targetElement.offsetTop - 80,
          behavior: 'smooth'
        });
      }
    });
  });
});
