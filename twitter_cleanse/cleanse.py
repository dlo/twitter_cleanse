#!/usr/bin/env python
# vim: set filetype=python

# Copyright 2015 Daniel Loewenherz
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import datetime
import errno
import hashlib
import json
import os
import pickle

import pytz
import tweepy


def get_twitter_oauth2(client_id, client_secret, redirect_uri="https://twitter.lionheartsw.com/callback", token_file="token.json"):
    if os.path.exists(token_file):
        with open(token_file) as f:
            token = json.load(f)
    else:
        token = None

    oauth2_user_handler = tweepy.OAuth2UserHandler(
        client_id=client_id,
        redirect_uri=redirect_uri,
        scope=[
            "users.read",
            "tweet.read",
            "follows.read", "follows.write",
            "list.read", "list.write",
            "offline.access",
        ],
        client_secret=client_secret
    )

    if token:
        oauth2_user_handler.token = token

    if not oauth2_user_handler.token:
        print("Please go to this URL and authorize the application:")
        print(oauth2_user_handler.get_authorization_url())

        authorization_response = input("Enter the full callback URL: ")

        oauth2_user_handler.fetch_token(authorization_response)

        with open(token_file, "w") as f:
            json.dump(oauth2_user_handler.token, f)

    return tweepy.Client(oauth2_user_handler)

CACHE_DIR = "twitter_cleanse_cache"

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def cached_request(client_method, use_cache, **kwargs):
    if not use_cache:
        return client_method(**kwargs)

    mkdir_p(CACHE_DIR)

    key_parts = [client_method.__name__]
    key_parts.extend(f"{k}={v}" for k, v in sorted(kwargs.items()))
    key = "".join(key_parts)
    checksum = hashlib.md5(key.encode('utf-8')).hexdigest()
    filename = os.path.join(CACHE_DIR, f"{checksum}.pickle")

    if os.path.exists(filename):
        try:
            with open(filename, 'rb') as f:
                return pickle.load(f)
        except (pickle.UnpicklingError, EOFError):
            # Cache file is corrupted, treat as a cache miss
            pass

    response = client_method(**kwargs)

    with open(filename, 'wb') as f:
        pickle.dump(response, f)

    return response


def cleanse(client, use_cache=True, years_dormant_threshold=2, dry_run=False):
    response = cached_request(client.get_me, use_cache)
    me = response.data
    user_id = me.id

    def get_list_id(name, description):
        response = cached_request(client.get_owned_lists, use_cache, id=user_id)
        lists = response.data or []

        for list_ in lists:
            if list_.name == name:
                break
        else:
            response = cached_request(
                client.create_list,
                use_cache,
                name=name,
                description=description,
                private=True
            )
            list_ = response.data

        return list_.id

    def unfollow_and_add_to_list(list_id, friend_user_id, screen_name):
        print(f"Unfollowing {screen_name} and adding to list.")
        if not dry_run:
            cached_request(client.add_list_member, use_cache, id=list_id, user_id=friend_user_id)
            cached_request(client.unfollow, use_cache, target_user_id=friend_user_id)

    # The 'Unfollowed: Quit Twitter' list is no longer used because
    # the v2 API doesn't easily support checking the last tweet date.
    # stopped_tweeting_list_id = get_list_id(
    #     name="Unfollowed: Quit Twitter",
    #     description="Users previously followed who have since stopped tweeting."
    # )

    no_tweets_list_id = get_list_id(
        name="Unfollowed: No Tweets",
        description="Users who have no tweets."
    )

    muted_list_id = get_list_id(
        name="Unfollowed: Muted",
        description="Muted users who are no longer followers."
    )

    follower_ids = set()
    pagination_token = None
    while True:
        response = cached_request(
            client.get_users_followers,
            use_cache,
            id=user_id,
            max_results=1000,
            pagination_token=pagination_token
        )
        if response.data:
            for user in response.data:
                follower_ids.add(user.id)
        
        if not response.meta or 'next_token' not in response.meta:
            break
        pagination_token = response.meta['next_token']

    pagination_token = None
    while True:
        response = cached_request(
            client.get_users_following,
            use_cache,
            id=user_id,
            max_results=1000,
            user_fields=["public_metrics", "muting"],
            pagination_token=pagination_token
        )

        if not response.data:
            if not response.meta or 'next_token' not in response.meta:
                break
            else:
                pagination_token = response.meta['next_token']
                continue

        for user in response.data:
            friend_user_id = user.id
            screen_name = user.username

            if user.muting and user.id not in follower_ids:
                unfollow_and_add_to_list(muted_list_id, friend_user_id, screen_name)
                print(f"Unfollowing {screen_name} since they're muted and are not a follower.")
            elif user.public_metrics['tweet_count'] == 0:
                unfollow_and_add_to_list(no_tweets_list_id, friend_user_id, screen_name)
                print(f"Unfollowing {screen_name} since they have no tweets")

        if not response.meta or 'next_token' not in response.meta:
            break
        pagination_token = response.meta['next_token']

