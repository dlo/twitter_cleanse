#!/usr/bin/env python

import os

try:
    from setuptools import setup
except ImportError:
    from distutils.core import setup

metadata = {}
execfile("twitter_cleanse/metadata.py", metadata)

setup(
    name='twitter_cleanse',
    version=metadata['__version__'],
    license=metadata['__license__'],
    description="Clean up your Twitter follow list.",
    author=metadata['__author__'],
    author_email=metadata['__email__'],
    url="http://dlo.me/",
    packages=[
        'twitter_cleanse',
    ],
    scripts=[
        'bin/twitter-cleanse',
    ],
    install_requires=[
        'python-dateutil',
        'restmapper',
        'requests-oauthlib',
        'pytz'
    ]
)

