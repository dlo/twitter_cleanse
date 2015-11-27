#!/usr/bin/env python
# vim: set filetype=python

import os
import json
import pprint
import datetime

import pytz
from dateutil import parser as date_parser
from restmapper.restmapper import RestMapper
from requests_oauthlib import OAuth1

Twitter = RestMapper("https://api.twitter.com/1.1/", url_transformer=lambda url: url + ".json")

def cleanse(consumer_key, consumer_secret, access_token, access_token_secret, use_cache, years_dormant_threshold=2, dry_run=False):
    auth = OAuth1(consumer_key, consumer_secret, access_token, access_token_secret)
    twitter = Twitter(auth=auth)

    def users():
        cursor = -1
        while cursor != 0:
            filename = "users_{}.json".format(cursor)
            if use_cache and os.path.exists(filename):
                with open(filename, 'r') as f:
                    result = json.load(f)
            else:
                result = twitter.friends.list(count=200, include_user_entities=False, cursor=cursor)

                if use_cache:
                    with open(filename, 'w') as f:
                        json.dump(result, f)

            for user in result['users']:
                yield user

            cursor = result['next_cursor']

    def retry_until_success(fun, **kwargs):
        while True:
            response = fun(
                parse_response=False,
                **kwargs
            )

            if response.status_code == 200:
                break
            else:
                print response.json()
                raw_input("{} status code returned. Wait 15 (?) minutes and then press enter.".format(response.status_code))

        return response.json()

    def get_list_id(name, description):
        lists = retry_until_success(twitter.lists.ownerships)['lists']

        # `list` is taken in Python
        for list_ in lists:
            if list_['name'] == name:
                break
        else:
            # If the loop falls through without finding a match, create a new list.
            list_ = twitter.POST.lists.create(name=name, mode="private", description=description)

        return list_['id']

    def unfollow_and_add_to_list(list_id, user_id, screen_name):
        if not dry_run:
            retry_until_success(
                twitter.POST.lists.members.create,
                list_id=stopped_tweeting_list_id,
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

    now = datetime.datetime.now(pytz.utc)
    for user in users():
        user_id = user['id']
        screen_name = user['screen_name']
        if 'status' in user:
            dt = date_parser.parse(user['status']['created_at'])
            delta = now - dt
            years = round(delta.days / 365., 2)
            if years >= years_dormant_threshold:
                unfollow_and_add_to_list(stopped_tweeting_list_id, user_id, screen_name)

                print "Unfollowing", screen_name, "since they haven't tweeted in {} years".format(years)
        else:
            unfollow_and_add_to_list(no_tweets_list_id, user_id, screen_name)

            print "Unfollowing", screen_name, "since they have no tweets"

