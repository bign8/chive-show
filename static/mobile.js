const BASE = '/api/v1/post?count=3'
var next_link // state updated by pump and init
let scroller = document.querySelector('.scroller')

const colors = ['primary', 'secondary', 'success', 'warning', 'danger', 'info', 'dark']

const signoffs = [
    `<h4 class="alert-heading">That's everything!</h4><p class="mb-0">Go for a walk.</p>`,
    `<h4 class="alert-heading">That's it!</h4><p class="mb-0">Better find a new hobby.</p>`,
    `<h4 class="alert-heading">Welp...</h4><p class="mb-0">You found the end of the internet!</p>`,
]

function random(list) {
    return list[Math.floor(Math.random() * list.length)]
}

function notify(msg, ...styles) {
    // https://getbootstrap.com/docs/5.0/components/alerts/
    let trailer = document.createElement('div')
    trailer.classList.add('alert', 'text-center', ...styles)
    trailer.innerHTML = msg
    scroller.append(trailer)
}

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
    if (rect.y > 0  && rect.y < window.innerHeight) e.target.play()
}

function create_media(media) {
    var img;
    if (media.url.endsWith(".mp4")) {
        let src = document.createElement('source')
        src.type = 'video/mp4'
        src.src = media.url
        img = document.createElement('video')
        img.loop = true
        img.muted = true // :shrug:
        img.disableRemotePlayback = true // experimental (don't show "cast" button on mobile)
        img.playsInline = true
        img.append(src)
        img.onloadeddata = play_if_visible
        img.preload = "metadata" // TODO: real lazy load; best we can do w/o custom logic on when to load content
        videos.observe(img) // have the scroll observer play if in viewport
    } else {
        img = document.createElement('img')
        img.src = media.url
        img.loading = "lazy"
        img.height = 400 // HACK: force browser to treat unloaded images with a reasonable height so not everything is loaded at once. would be awesome if server provided image dimensions
        img.onload = img.removeAttribute.bind(img, 'height')
    }

    if (media.title) {
        img.title = media.title
        img.alt = media.title
    }

    let card = document.createElement('div')
    card.classList.add('card')
    card.append(img)

    if (media.caption) {
        let body = document.createElement('div')
        body.classList.add('card-body')
        card.append(body)
        body.innerHTML = media.caption
        for (let child of body.children) child.classList.add('card-text')
    } else {
        img.classList.add('card-img-bottom')
    }
    img.classList.add('card-img-top')
    return card
}

function create_tag(tag) {
    // TODO: build user list of colored categories (and defaults)
    // TODO: create category equivalency map (ie: DAR = Daily Afternoon ...)
    let color = random(colors)

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
    mug.alt = post.creator
    mug.title = post.creator
    mug.loading = "lazy"
    banner.append(mug)

    let since = document.createElement('small')
    since.innerText = timeSince(new Date(post.date))
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
    if (!next_link) return notify(random(signoffs), 'alert-warning')
    fetch(next_link).then(r => r.json()).then(res => {
        next_link = res.next_url
        if (!res.self_url) return notify(
            `<h4 class="alert-header">Whoa</h4><p class="mb-0">We didn't find anything!<br/>Try <a href="#">Resetting filters</a></p>`,
            'alert-danger', 'position-absolute', 'top-50', 'start-50', 'translate-middle'
        )
        if (!next_link && !res.data) return pump() // if we hit the end perfectly
        return res.data
    }).then(posts => {
        if (!posts) return
        for (let post of posts) scroller.append(create_post(post))
        bottom.observe(scroller.lastChild.previousSibling ?? scroller.lastChild)
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
    scroller.innerHTML = ''
    next_link = BASE
    let tag = document.location.hash.slice(1)
    if (tag) next_link += "&tag=" + tag
    pump()
}

// let's do this!
window.addEventListener('hashchange', init)
init()

// Tag selector options (https://getbootstrap.com/docs/5.0/components/offcanvas/)
let tags = document.getElementById('tags')
let bs_tags = new bootstrap.Offcanvas(tags)
function init_offscreen() {
    fetch('/api/v1/tags').then(r => r.json()).then(data => {
        let list = document.querySelector('.list-group')
        list.innerHTML = ''
        for (const [tag, count] of Object.entries(data.tags)) {
            let ele = document.createElement('a')
            ele.classList.add('list-group-item', 'd-flex', 'justify-content-between', 'align-items-center')
            ele.innerText = tag
            ele.href = '#' + tag
            ele.addEventListener('click', e => bs_tags.hide())
            list.append(ele)

            let span = document.createElement('span')
            span.classList.add('badge', 'bg-secondary', 'rounded-pill') // TODO: consistent coloring
            span.style.width = '3em'
            span.innerText = count
            ele.append(span)
        }
        tags.removeEventListener('show.bs.offcanvas', init_offscreen)
    })
}
tags.addEventListener('show.bs.offcanvas', init_offscreen)