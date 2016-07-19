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
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Environment :: Console",
        "Intended Audience :: End Users/Desktop",
        "License :: OSI Approved :: Apache Software License",
        "Natural Language :: English",
        "Natural Language :: English",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 2.6",
        "Programming Language :: Python :: 2.7",
        "Programming Language :: Python",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Utilities",
    ],
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

