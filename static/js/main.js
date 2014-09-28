"use strict";
// TODO: resize handler (adjust margins and image size)
// TODO: implement videos
// TODO: implement settings page / login / logout / app page (angular.js?)
// TODO: local weather: http://erikflowers.github.io/weather-icons/
// TODO: filter "best links of the internet" posts
// TODO: disable next button while loading next
// TODO: remove jquery: http://youmightnotneedjquery.com/

/* Front End Documentation:
http://bootswatch.com/cyborg/
http://www.html5rocks.com/en/tutorials/es6/promises/
*/

var Chive = {};

// TIMEOUT INDICATOR (in progres)
Chive.loader = function () {
    // thanks: http://jakearchibald.com/2013/animated-line-drawing-svg/
    // thanks: http://tympanus.net/codrops/2014/04/09/how-to-create-a-circular-progress-button/


    var svg, path, anim, len;

    var load = function (settings) {

        // TODO: provide defaults
        settings = settings || {};
        settings.diameter = 100;
        settings.strokeWidth = 4;
        settings.color = '#02fd00';
        settings.parent = $('.next .container')[0]

        // setup variables
        var ns = 'http://www.w3.org/2000/svg';
        var d = 'M c c m -r, 0 a r,r 0 1,1 d,0 a r,r 0 1,1 -d,0'; // thanks: http://stackoverflow.com/a/10477334
        var circle_diameter = settings.diameter - settings.strokeWidth;

        // Adjust Path String
        d = d.replace(/c/g, settings.diameter / 2);
        d = d.replace(/d/g, circle_diameter);
        d = d.replace(/r/g, circle_diameter / 2);

        // Create SVG Element
        svg = document.createElementNS(ns, 'svg');
        svg.setAttribute('height', settings.diameter);
        svg.setAttribute('width', settings.diameter);

        // Create PATH Element
        path = document.createElementNS(ns, 'path');
        path.setAttribute('d', d);
        path.style.stroke = settings.color;
        path.style.strokeWidth = settings.strokeWidth;
        path.style.transition = path.style.WebkitTransition = 'stroke-dashoffset 2s linear';
        len = path.getTotalLength();
        path.style.strokeDasharray = len + ' ' + len;
        path.style.strokeDashoffset = len;
        svg.appendChild(path);

        // Rotate dash start indefinitely
        anim = document.createElementNS(ns, 'animateTransform');
        anim.setAttribute('attributeName', 'transform');
        anim.setAttribute('type', 'rotate');
        anim.setAttribute('from', '0 c c'.replace(/c/g, settings.diameter / 2));
        anim.setAttribute('to', '360 c c'.replace(/c/g, settings.diameter / 2));
        anim.setAttribute('dur', 8);
        anim.setAttribute('repeatCount', 'indefinite');
        path.appendChild(anim);

        // Append to document
        settings.parent.appendChild(svg);
        return function (percent) {
            // can pass in negative numbers to animate backwards
            path.style.strokeDashoffset = len * (1 - percent);
        };
    };

    return load;
}();


Chive.ajax = function () {
    function ajax(method, url, data) {
        var promise = new Promise(function (accept, reject) {
            var req = new XMLHttpRequest();
            req.open(method, url, true);
            req.onload = function () {
                if (req.status >= 200 && req.status < 400) {
                    accept(req.responseText);
                } else
                    reject();
            }
            req.onerror = reject.bind(this);
            if (data) req.setRequestHeader(
                'Content-Type',
                'application/x-www-form-urlencoded; charset=UTF-8'
            );
            req.send(data);
        });
        return promise.then(JSON.parse); // assumed JSON
    }

    return {
        'get': ajax.bind(this, 'GET'),
        'put': ajax.bind(this, 'PUT'),
        'post': ajax.bind(this, 'POST'),
        'delete': ajax.bind(this, 'DELETE'),
        // jsonp: https://github.com/bign8/bign8.github.io/blob/master/404.js#L53
    }
}();

Chive.date_format = function () {
    // TODO: Improve this!
    var months_short = [
        'Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'
    ];

    return function(date) {
        var str = months_short[date.getMonth()] + ' ';
        str += date.getDate() + ', ';
        str += date.getFullYear() + '&emsp;';
        str += date.toLocaleTimeString("en-US", {hour: 'numeric', minute: 'numeric'});
        return str;
    };
}();

Chive.shuffle = function (array) {
    // Shuffle locally (src: http://bost.ocks.org/mike/shuffle/)
    var m = array.length, t, i;
    while (m) { // While there remain elements to shuffle…
        i = Math.floor(Math.random() * m--); // Pick a remaining element
        t = array[m]; // And swap it with the current element
        array[m] = array[i];
        array[i] = t;
    }
    return array;
};

Chive.ucfirst = function(str) {
    str += '';
    var f = str.charAt(0).toUpperCase();
    return f + str.substr(1).toLowerCase();
};

Chive.get_cat_label = function () {
    // TODO: build user list of colored categories (and defaults)
    // TODO: create category equivalency map (ie: DAR = Daily Afternoon ...)
    var labels = ['default', 'primary', 'success', 'warning', 'danger', 'info'];

    return function (cat) {
        return labels[Math.floor(Math.random() * labels.length)];
    };
}();

Chive.timeout = function () {
    var timer = function (fn, timer) {
        this.active = true;

        var id = setTimeout((function () {
            this.clear();
            fn.call();
        }).bind(this), timer);

        this.clear = function () {
            clearTimeout(id);
            id = null;
            this.active = false;
            this.clear = function () {};
        };
    };

    return function (fn, time) {
        return new timer(fn, time);
    };
}();

Chive.viewer = function () {
    // TODO: assignable default interval

    // Local Variables
    var list = [];
    var active = null;
    var item = null;
    var timeout = null;

    // (Re)load list from server
    function load_list() {
        return Chive.ajax.get('/api/post/random').then(function (res) {
            if (res.status != 'success') throw Error(res.data);
            list = Chive.shuffle(res.data);
        });
    }

    // Prefetch animage and return the associated promise
    function fetch_image(url) {
        // TODO: move off to own package
        // TODO: Timeout spiral for next image
        //       http://tympanus.net/codrops/2014/04/09/how-to-create-a-circular-progress-button/
        //       http://jakearchibald.com/2013/animated-line-drawing-svg/
        // TODO: progress of image downloading (show progress spiral if it takes more than 1 sec)
        //       http://blogs.adobe.com/webplatform/2012/01/13/html5-image-progress-events/
        var promise = new Promise(function(accept, reject) {
            var image = new Image();
            image.addEventListener('load', accept.bind(this, image));
            image.addEventListener('error', reject.bind(this));
            image.src = url;
        });
        return promise;
    }

    // Render current category
    function render_active_category() {
        active.media = Chive.shuffle(active.media);
        console.log(active);

        // Print title
        var title = active.title.replace(/\([^\)]*\)/, '').trim();
        $('#feed-title').html(title);

        // Print Categories
        var cat, label, name, ele = $('#categories').empty();
        for (var i in active.tags) {
            cat = active.tags[i];
            label = Chive.get_cat_label(cat);
            name = Chive.ucfirst(cat);
            ele.append('&nbsp;<span class="label label-' + label + '">' + name + '</span>');
        }

        // Date Published
        var published = new Date(active.date);
        $('#posted').html('<small>' + Chive.date_format(published) + '</small>');
    }

    // Change image in application
    function run() {

        // If list is empty -> reload
        if (!list.length)
            return reload_list().then(run);

        // Check if curent article is active
        if (!active || !active.media.length) {
            active = list.shift();
            render_active_category();
        }

        render_next_item();
    }

    // Render image to dom
    function render_next_item() {
        // TODO: fade images through: http://api.jquery.com/fadeout/
        item = active.media.shift();
        var promise = fetch_image(item.url).then(function(image) {
            var frame = '#frame-1';

            // Scale Image for height
            var height = $(document).height() - 105; // header + footer + 3
            if (image.height > height) {
                image.width *= height / image.height;
                image.height = height;
            }

            // Replace content of image
            var margin = -Math.round(image.height / 2);
            $('.push', frame).css('marginBottom', margin + 'px');
            $('.container', frame).empty().append(image);
        }, function() {
            // TODO: callback to server with info and mark as an error occured
        }).then(function () {
            timeout = Chive.timeout(run, 20*1e3);
        });
        return promise;
    }

    $('#btn-next').click(function () {
        timeout.clear();
        run();
    });

    return {
        'pause': function () {
            timeout.clear();
        },
        'start': function () {
            if (!timeout || !timeout.active) {
                var promise = list.length ? Promise.resolve() : load_list();
                promise.then(run);
            }
        },
    };
}();

$(document).ready(Chive.viewer.start);
