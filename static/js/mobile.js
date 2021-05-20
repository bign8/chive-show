let next_link = '/api/v1/post/random?count=10'
let scroller = document.querySelector('.scroller')

// convert json data into html data
function create_post(post) {
    let div = document.createElement('div')
    div.classList.add('post')

    let title = document.createElement('h5')
    title.innerText = post.title
    title.title = post.title
    div.append(title)

    let pre = document.createElement('pre')
    pre.innerText = JSON.stringify(post, null, 2)
    div.append(pre)

    return div
}

// load the next page of posts
function pump() {
    fetch(next_link).then(r => r.json()).then(res => {
        // TODO: listen to server's response and stream through posts this way
        return res.data
    }).then(posts => {
        for (let post of posts) scroller.append(create_post(post))
        observer.observe(scroller.lastChild.previousSibling)
    });
}

// https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
let observer = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            observer.unobserve(entry.target)
            pump()
        }
        console.log(entry)
    })
})

// let's do this!
pump()