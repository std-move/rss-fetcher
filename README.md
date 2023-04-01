# RSS fetcher

This repository currently provides a feed of "bigTop" WSJ articles sourced from their World News feed. WSJ's Big Top articles are usually well researched and interesting, however there is no official feed for them. The feed is created by fetching WSJ's World News RSS feed and filtering it, leaving in only these in-depth articles. Only data that is in the original feed gets into the filtered feed - no scraping/article text extraction is attempted.

Despite the source feed name, as of March 2023 it started providing all kinds of articles, not just ones from their World News section. We attempt to filter the feed for news articles only.

The feed is present in branch gh-pages and is updated every hour. Permalink: https://std-move.github.io/rss-fetcher/wsj-world-bigtop.xml
