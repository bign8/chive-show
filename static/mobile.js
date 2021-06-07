const BASE = '/api/v1/post?count=3'
var next_link // state updated by pump and init
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

function play_if_visible(e) {
    let rect = e.target.getBoundingClientRect()
    if (rect.y > 0  && rect.y < window.innerHeight) {
        e.target.play()
        console.log('to-play', e.target)
    }
}

function create_media(media) {
    // TODO: lazy load media (initial load is pretty heavy)

    var img;
    if (media.url.endsWith(".mp4")) {
        let src = document.createElement('source')
        src.type = 'video/mp4'
        src.src = media.url
        img = document.createElement('video')
        img.loop = true
        img.muted = true // :shrug:
        img.playsInline = true
        img.append(src)
        img.onloadeddata = play_if_visible
        videos.observe(img) // have the scroll observer play if in viewport
    } else {
        img = document.createElement('img')
        img.src = media.url
    }

    let card = document.createElement('div')
    card.classList.add('card')
    img.classList.add('card-img-top')
    card.append(img)

    if (media.title) {
        let body = document.createElement('div')
        body.classList.add('card-body')
        card.append(body)

        let text = document.createElement('small')
        text.classList.add('card-text')
        img.title = media.title
        img.alt = media.title
        text.innerText = media.title
        body.append(text)
    } else {
        img.classList.add('card-img-bottom')
    }
    return card
}

function create_tag(tag) {
    // TODO: build user list of colored categories (and defaults)
    // TODO: create category equivalency map (ie: DAR = Daily Afternoon ...)
    const colors = ['primary', 'secondary', 'success', 'warning', 'danger', 'info', 'dark']
    let color = colors[Math.floor(Math.random() * colors.length)]

    let badge = document.createElement('a')
    badge.href = '#' + tag
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

    let banner = document.createElement('div')
    banner.classList.add('banner')
    div.append(banner)

    let title = document.createElement('h5')
    title.innerText = post.title
    title.title = post.title
    title.addEventListener('click', e => {
        window.open(post.link)
    })
    banner.append(title)

    let mug = document.createElement('img')
    mug.src = post.mugshot
    mug.style.height = '20px'
    mug.style.width = '20px'
    mug.classList.add('rounded-circle')
    banner.append(mug)

    let since = document.createElement('small')
    since.innerText = post.creator + ' ' + timeSince(new Date(post.date))
    since.style.marginLeft = '.25em'
    since.classList.add('text-muted')
    banner.append(since);

    for (let tag of post.tags) banner.append(create_tag(tag))

    let wrap = document.createElement('div')
    wrap.classList.add('bar-wrap')
    banner.append(wrap)

    let bar = document.createElement('div')
    bar.classList.add('bar')
    wrap.append(bar)

    for (let media of post.media) div.append(create_media(media))

    // let pre = document.createElement('pre')
    // pre.innerText = JSON.stringify(post, null, 2)
    // pre.style.marginBottom = '0'
    // div.append(pre)

    progress.observe(div)
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
        if (entry.target.readyState != entry.target.HAVE_ENOUGH_DATA) return // let's not try and do things while content is loading
        if (entry.isIntersecting && entry.target.paused) {
            try {
                entry.target.play()
                entry.target.parentElement.classList.remove('giflock')
            } catch (error) {
                console.log('cannot play w/o user input?', error)
                entry.target.parentElement.classList.add('giflock')
            }
        } else if (!entry.target.paused) {
            entry.target.pause()
        }
    })
})
let progress = new IntersectionObserver((entries, observer) => {
    // entries.forEach(entry => {
    //     console.log(entry)
    //     // TODO: this notifies when things enter and leave viewport
    //     // Use this to manage the set of posts that need to have there title-bar progress updated
    // })
})

// update progress bars on scroll
// source: https://www.w3schools.com/howto/howto_js_scroll_indicator.asp
function updateProgress() {
    document.querySelectorAll('.post').forEach(node => {
        let x = node.getBoundingClientRect()
        node.querySelector('.bar').style.width = (x.top / -x.height * 100).toString() + "%"
    })
}
window.addEventListener('scroll', e => window.requestAnimationFrame(updateProgress))

function init() {
    console.log('about to init')
    scroller.innerHTML = ''
    next_link = BASE
    let tag = document.location.hash.slice(1)
    if (tag) next_link += "&tag=" + tag
    pump()
}

// let's do this!
window.addEventListener('hashchange', init)
init()