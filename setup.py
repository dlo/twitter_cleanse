#!/usr/bin/env python
# -*- coding: utf-8 -*-

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

import os

try:
    from setuptools import setup
except ImportError:
    from distutils.core import setup

def read(fname):
    return open(os.path.join(os.path.dirname(__file__), fname)).read()

metadata = {}
metadata_file = "twitter_cleanse/metadata.py"
exec(compile(open(metadata_file).read(), metadata_file, 'exec'), metadata)

classifiers = [
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
]

setup(
    name='twitter_cleanse',
    classifiers=classifiers,
    version=metadata['__version__'],
    license=metadata['__license__'],
    description="Clean up your Twitter follow list.",
    author=metadata['__author__'],
    author_email=metadata['__email__'],
    url="https://github.com/dlo/twitter_cleanse",
    packages=['twitter_cleanse'],
    package_data={'': ['LICENSE', 'README.md']},
    scripts=['bin/twitter-cleanse'],
    install_requires=read("requirements.txt").split("\n")
)

