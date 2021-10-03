# [theChive](http://thechive.com) Show
[![Actions](https://github.com/bign8/chive-show/workflows/Go/badge.svg)](https://github.com/bign8/chive-show/actions) [![codecov.io](http://codecov.io/github/bign8/chive-show/coverage.svg)](http://codecov.io/github/bign8/chive-show)

*A collection of your favorite Chive images (taylored specifically for you)*

Ever want a safe for work slideshow of the images from *"Probably the best site in the World"*?  Well now you can enjoy these amazing images without the distractions when viewing them within their original context.

# The obligatory questions
As with any open source project, there are always someone wondering what the backstory of the project is.  These are my attempts to answer those questions.

## *Who* are you?
Just another developer online, looking to better himself through the growth of their nerd experiences.  If you would like to learn all about me, you are welcome to investigate my [personal website](http://bign8.info) at your own leisure.

## *What* the heck do you think your doing?
Having fun!  This is not meant to take away from theChive in any way.  If you have issues with this project, please contact me out of band and I will see what we can agree upon.

## *Where* are you building this?
Currently, this application is build to run on Google App Engine.  While this may be painful for some, their free tier allows quite a bit of versatility.  This project was designed to see if I can learn this new service platform.

## *When* are you building this?
This project is being developed on my personal time for no pay other than the general improvement of my architect and developer skills.  If you like what you see here, image what I can do if you pay me to work on a project.

## *Why* are you doing this?
I enjoy learning new technologies and this project allows me to tinker with several languages/techniques that I have never worked with before.  A few of the languages in play are as follows.

- Google App Engine
- Go lang
- 100% test coverage (hell no)
- CI runs all the time
- Lots of good API things
  - [HATEOAS](https://en.wikipedia.org/wiki/HATEOAS)
  - REST ish, but not
  - More ... http://www.nurkiewicz.com/2015/07/restful-considered-harmful.html
- Etc...

## *How* can I contribute?
Go ahead and fork the repository, create a branch with your changes and make a PR back to my fork.  Please include a description of why the change is necessary, and how you sought to fix the issue.

# To Do List
This list is extremely lengthy and will most likely never get done.

## User Accounts
It would be nice if users could login and save preferences.  This will also track user sessions and not show repeated images and articles.  This involves smart random selections, potentially targeted toward certain tags selected during a viewing preferences stage.

## Cloud Sourced Rating and Usability
Allow users to up-vote certain images and create user specific experience.  Tags associated with an image can receive be voted upon. Use running mean to keep scores normalized and more lenient toward recent votes.  Votes can be stored individually and processed once a day.

- [Running Average](https://en.wikipedia.org/wiki/Moving_average)

## Include Post Metadata + Dynamic Content
Expand application support to also show Image descriptions when available.  The support of video content should also be enabled, but be able to be toggled via a user account setting or one of the options during the experience setup phase.

## Streamline and Sexify UI and UX
Currently, the application starts up very crudely and is hard on the eyes.  Make the application start in a options selection mode where users can choose tags that look interesting and to support videos or not (sound settings too).  The timer on the icon (the cool green rotating svg thing) should be synced to the actual timing of image changes.  Images should be pre-buffered and fade in and out too.  Lots of possible work here.
