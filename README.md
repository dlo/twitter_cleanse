Twitter Cleanse
===============

How many people do you follow on Twitter? If it's a fairly large number (as in my case), there's a good chance that a few of them don't actually tweet anymore. Because I'm not a huge fan of clutter—digital or otherwise—I decided to write a utility to clean up the list of people I follow and only keep following those users who've kept up their Twitter habit.

In short, this utility unfollows everyone who you follow who:

1. Has no tweets, or
2. Hasn't tweeted in the last X years (defaults to 2).
3. Has been muted but no longer follows you back.

It will also save these users to distinct Twitter lists in case you'd like to re-follow them in the future.

Installation
------------

    pip install python-dateutil restmapper requests-oauthlib pytz twitter-cleanse

Running
-------

Before you start anything, you'll need to create a [Twitter application](https://apps.twitter.com). Once you have all the credentials you need (consumer key, consumer secret, access token, and access token secret), just view the built-in command help for usage.

    twitter-cleanse -h

What it does
------------

When you run the script with all the required credentials, it goes through every person you follow and checks if they have any tweets. If they don't, they get unfollowed and added to a private list called "Unfollowed: No Tweets" (in case you want to re-follow them in the future). If they _have_ tweeted, the date of their last tweet is compared against a threshold (defaults to 2 years). If their last tweet was made further in the past than the threshold, then they are unfollowed and also placed into a (separate) list called "Unfollowed: Quit Twitter".

Contributing
------------

Feel free to send a pull request! I can't promise to accept everything, but if your addition fixes a bug or adds a cool new feature, it'll most likely make it in.

TODOs
-----

* Tests
* ~~Dry run~~
