let next_link = '/api/v1/post/random?count=10'
let scroller = document.querySelector('.scroller')

// https://stackoverflow.com/a/3177838
function timeSince(date) {
    var seconds = Math.floor((new Date() - date) / 1000);
    var interval = seconds / 31536000;
    if (interval > 2) {
      return Math.floor(interval) + " years";
    }
    interval = seconds / 2592000;
    if (interval > 2) {
      return Math.floor(interval) + " months";
    }
    interval = seconds / 86400;
    if (interval > 2) {
      return Math.floor(interval) + " days";
    }
    interval = seconds / 3600;
    if (interval > 2) {
      return Math.floor(interval) + " hours";
    }
    interval = seconds / 60;
    if (interval > 2) {
      return Math.floor(interval) + " minutes";
    }
    return Math.floor(seconds) + " seconds";
}

function create_media(media) {
    var img;
    if (media.url.endsWith(".mp4")) {
        let src = document.createElement('source')
        src.type = 'video/mp4'
        src.src = media.url
        img = document.createElement('video')
        img.loop = true
        img.playsInline = true
        img.append(src)
        videos.observe(img) // have the scroll observer play if in viewport
    } else {
        img = document.createElement('img')
        img.src = media.url
        img.title = media.title
        img.alt = media.title
    }

    // dumb huristic to determin if the title is human readable
    if (img.title.indexOf(' ') > 0) {
        let card = document.createElement('div')
        card.classList.add('card')
        card.classList.add('card-img-top')
        card.append(img)

        let body = document.createElement('div')
        body.classList.add('card-body')
        card.append(body)

        let text = document.createElement('small')
        text.classList.add('card-text')
        text.innerText = media.title
        body.append(text)

        card.style.marginTop = '1em'
        card.style.marginBottom = '1em'
        return card
    }

    img.classList.add('img-fluid')
    img.classList.add('mx-auto')
    img.classList.add('d-block')
    img.style.marginTop = '1em'
    img.style.marginBottom = '1em'
    return img
}

function create_tag(tag) {
    // TODO: build user list of colored categories (and defaults)
    // TODO: create category equivalency map (ie: DAR = Daily Afternoon ...)
    const colors = ['primary', 'secondary', 'success', 'warning', 'danger', 'info', 'dark']
    let color = colors[Math.floor(Math.random() * colors.length)]

    let badge = document.createElement('span')
    badge.innerText = tag
    badge.classList.add('badge')
    badge.classList.add('bg-' + color)
    badge.classList.add('float-end')
    badge.style.marginRight = '.25em'
    badge.style.marginBottom = '.25em'
    return badge
}

// convert json data into html data
function create_post(post) {
    let div = document.createElement('div')
    div.classList.add('post')
    div.addEventListener('click', e => {
        window.open(post.link)
    })

    let banner = document.createElement('div')
    div.append(banner)

    let title = document.createElement('h5')
    title.innerText = post.title
    title.title = post.title
    banner.append(title)

    let mug = document.createElement('img')
    mug.src = post.mugshot
    mug.style.height = '20px'
    mug.style.width = '20px'
    mug.classList.add('rounded-circle')
    banner.append(mug)

    let since = document.createElement('small');
    since.innerText = post.creator + ' ' + timeSince(new Date(post.date))
    since.style.marginLeft = '.25em'
    since.classList.add('text-muted')
    banner.append(since);

    for (let tag of post.tags) banner.append(create_tag(tag))

    for (let media of post.media) div.append(create_media(media))

    // let pre = document.createElement('pre')
    // pre.innerText = JSON.stringify(post, null, 2)
    // pre.style.marginBottom = '0'
    // div.append(pre)

    return div
}

// load the next page of posts
function pump() {
    if (!next_link) {
        alert("Thats everyting! Go for a walk!")
        return
    }
    fetch(next_link).then(r => r.json()).then(res => {
        next_link = res.next_url
        return res.data
    }).then(posts => {
        for (let post of posts) scroller.append(create_post(post))
        bottom.observe(scroller.lastChild.previousSibling)
    })
}

// https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
let bottom = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            observer.unobserve(entry.target)
            pump()
        }
    })
})
let videos = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
        if (entry.isIntersecting && entry.target.paused) {
            console.log('about to play video', entry)
            entry.target.play()
        } else if (!entry.target.paused) {
            console.log('about to pause video', entry)
            entry.target.pause()
        }
    })
})

// let's do this!
pump()