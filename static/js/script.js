const posts = document.querySelectorAll('.post-toggle');

posts.forEach(post => {
    post.addEventListener('click', () => {
        // removeActivePosts()
        post.parentNode.classList.toggle('active')
        post.classList.toggle('flip')
    })
})

function removeActivePosts() {
    post.parentNode.classList.remove('active')
    post.classList.remove('flip')

    // posts.forEach(post => {
    //     if (post.classList.contains('flip')) {
    //     }
    // })
}
