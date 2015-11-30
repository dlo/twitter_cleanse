#!/usr/bin/env python
# vim: set filetype=python

import os
import json
import pprint
import datetime
import hashlib
import urllib
import errno

import pytz
from dateutil import parser as date_parser
from restmapper.restmapper import RestMapper
from requests_oauthlib import OAuth1

def url_transformer(url):
    if url.endswith("/__name__"):
        url = url.replace("/__name__", "")

    return url + ".json"


def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc: # Python >2.5
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else: raise


Twitter = RestMapper("https://api.twitter.com/1.1/", url_transformer=url_transformer)


def request_hash(fun, **params):
    path = "/".join(fun.components)
    key = "{}{}".format(path, urllib.urlencode(params))
    checksum = hashlib.md5(key).hexdigest()
    return checksum


def cleanse(consumer_key, consumer_secret, access_token, access_token_secret, use_cache, years_dormant_threshold=2, dry_run=False):
    auth = OAuth1(consumer_key, consumer_secret, access_token, access_token_secret)
    twitter = Twitter(auth=auth)

    def retry_until_success(fun, **kwargs):
        mkdir_p("twitter_cleanse_cache")
        filename = "twitter_cleanse_cache/{}.json".format(request_hash(fun, **kwargs))
        if use_cache and os.path.exists(filename):
            with open(filename, 'r') as f:
                result = json.load(f)
        else:
            while True:
                response = fun(
                    parse_response=False,
                    **kwargs
                )

                if response.status_code == 200:
                    result = response.json()
                    if use_cache:
                        with open(filename, 'w') as f:
                            json.dump(result, f)

                    break
                else:
                    raw_input("{} status code returned. Wait 15 (?) minutes and then press enter.".format(response.status_code))

        return result

    def users(fun):
        cursor = -1
        while cursor != 0:
            result = retry_until_success(
                fun,
                count=200,
                include_user_entities="false",
                cursor=cursor
            )

            for user in result['users']:
                yield user

            cursor = result['next_cursor']


    def get_list_id(name, description):
        lists = retry_until_success(twitter.lists.ownerships)['lists']

        # `list` is taken in Python
        for list_ in lists:
            if list_['name'] == name:
                break
        else:
            # If the loop falls through without finding a match, create a new list.
            list_ = retry_until_success(
                twitter.POST.lists.create,
                name=name,
                mode="private",
                description=description
            )

        return list_['id']

    def unfollow_and_add_to_list(list_id, user_id, screen_name):
        if not dry_run:
            retry_until_success(
                twitter.POST.lists.members.create,
                list_id=list_id,
                user_id=user_id,
                screen_name=screen_name
            )

            retry_until_success(
                twitter.POST.friendships.destroy,
                user_id=user_id
            )

    stopped_tweeting_list_id = get_list_id(
        name="Unfollowed: Quit Twitter",
        description="Users previously followed who have since stopped tweeting."
    )

    no_tweets_list_id = get_list_id(
        name="Unfollowed: No Tweets",
        description="Users who have no tweets."
    )

    muted_list_id = get_list_id(
        name="Unfollowed: Muted",
        description="Muted users who are no longer followers."
    )

    now = datetime.datetime.now(pytz.utc)
    followers = set(user['id'] for user in users(twitter.followers.list))

    for user in users(twitter.friends.list):
        user_id = user['id']
        screen_name = user['screen_name']
        muted = user['muting']
        if muted and user_id not in followers:
            unfollow_and_add_to_list(muted_list_id, user_id, screen_name)

            print "Unfollowing", screen_name, "since they're muted and are not a follower."
        elif 'status' in user:
            # Unfollow users who haven't tweeted in `years_dormant_threshold` years.
            dt = date_parser.parse(user['status']['created_at'])
            delta = now - dt
            years = round(delta.days / 365., 2)
            if years >= years_dormant_threshold:
                unfollow_and_add_to_list(stopped_tweeting_list_id, user_id, screen_name)

                print "Unfollowing", screen_name, "since they haven't tweeted in {} years".format(years)
        else:
            # Unfollow users who have never tweeted
            unfollow_and_add_to_list(no_tweets_list_id, user_id, screen_name)

            print "Unfollowing", screen_name, "since they have no tweets"

