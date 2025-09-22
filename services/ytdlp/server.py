#!/usr/bin/env python3
"""
yt-dlp HTTP service for Discord music bot
Provides a JSON API for video extraction and search functionality
"""

import asyncio
import json
import logging
import os
import sys
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Any

import aiohttp
from aiohttp import web, ClientSession
import yt_dlp
from concurrent.futures import ThreadPoolExecutor
import argparse


class YTDLPService:
    """HTTP service wrapper for yt-dlp functionality"""

    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.executor = ThreadPoolExecutor(max_workers=config.get('max_workers', 4))
        self.start_time = datetime.now()
        self.request_count = 0
        self.error_count = 0
        self.cache = {}
        self.cache_ttl = timedelta(hours=config.get('cache_ttl_hours', 24))

        # Create cache directory
        cache_dir = Path(config.get('cache_dir', '/tmp/ytdlp-cache'))
        cache_dir.mkdir(parents=True, exist_ok=True)

        # Setup logging
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
        self.logger = logging.getLogger('ytdlp-service')

        # Common yt-dlp options
        self.ytdl_opts = {
            'format': config.get('format', 'bestaudio/best'),
            'noplaylist': True,
            'ignoreerrors': True,
            'no_warnings': False,
            'extractaudio': True,
            'audioformat': config.get('audio_format', 'opus'),
            'audioquality': config.get('audio_quality', '128'),
            'outtmpl': str(cache_dir / '%(id)s.%(ext)s'),
            'restrictfilenames': True,
            'logtostderr': False,
            'quiet': True,
        }

        # Rate limiting options
        if config.get('rate_limit'):
            self.ytdl_opts['ratelimit'] = config['rate_limit']
        if config.get('sleep_interval'):
            self.ytdl_opts['sleep_interval'] = config['sleep_interval']

    async def health_check(self, request):
        """Health check endpoint"""
        uptime = datetime.now() - self.start_time
        return web.json_response({
            'success': True,
            'data': {
                'status': 'healthy',
                'version': yt_dlp.version.__version__,
                'uptime': str(uptime),
                'request_count': self.request_count,
                'error_count': self.error_count,
                'worker_count': self.config.get('max_workers', 4),
                'last_check': datetime.now().isoformat()
            }
        })

    async def extract_info(self, request):
        """Extract video information from URL"""
        try:
            data = await request.json()
            url = data.get('url')

            if not url:
                return web.json_response({
                    'success': False,
                    'error': 'URL is required',
                    'code': 400
                }, status=400)

            self.request_count += 1

            # Check cache first
            cache_key = f"extract:{url}"
            if cache_key in self.cache:
                cache_entry = self.cache[cache_key]
                if datetime.now() - cache_entry['timestamp'] < self.cache_ttl:
                    self.logger.info(f"Cache hit for URL: {url}")
                    return web.json_response({
                        'success': True,
                        'data': cache_entry['data']
                    })

            # Extract info using yt-dlp
            loop = asyncio.get_event_loop()
            info = await loop.run_in_executor(
                self.executor,
                self._extract_info_sync,
                url,
                data.get('format')
            )

            if info:
                # Cache the result
                self.cache[cache_key] = {
                    'data': info,
                    'timestamp': datetime.now()
                }

                return web.json_response({
                    'success': True,
                    'data': info
                })
            else:
                self.error_count += 1
                return web.json_response({
                    'success': False,
                    'error': 'Failed to extract video information',
                    'code': 404
                }, status=404)

        except Exception as e:
            self.error_count += 1
            self.logger.error(f"Error extracting info: {str(e)}")
            return web.json_response({
                'success': False,
                'error': str(e),
                'code': 500
            }, status=500)

    def _extract_info_sync(self, url: str, format_override: Optional[str] = None) -> Optional[Dict]:
        """Synchronous video info extraction"""
        try:
            opts = self.ytdl_opts.copy()
            if format_override:
                opts['format'] = format_override

            with yt_dlp.YoutubeDL(opts) as ydl:
                info = ydl.extract_info(url, download=False)

                if not info:
                    return None

                # Clean and structure the info
                clean_info = {
                    'id': info.get('id', ''),
                    'title': info.get('title', ''),
                    'description': info.get('description', ''),
                    'duration': info.get('duration'),
                    'webpage_url': info.get('webpage_url', url),
                    'thumbnail': self._get_best_thumbnail(info.get('thumbnails', [])),
                    'uploader': info.get('uploader', ''),
                    'upload_date': info.get('upload_date', ''),
                    'view_count': info.get('view_count'),
                    'extractor': info.get('extractor', ''),
                    'extractor_key': info.get('extractor_key', ''),
                    'available': True,
                    'live_status': info.get('live_status'),
                    'tags': info.get('tags', []),
                    'categories': info.get('categories', []),
                    'formats': self._clean_formats(info.get('formats', [])),
                    'thumbnails': self._clean_thumbnails(info.get('thumbnails', []))
                }

                return clean_info

        except Exception as e:
            self.logger.error(f"yt-dlp extraction error for {url}: {str(e)}")
            return None

    async def search(self, request):
        """Search for videos"""
        try:
            data = await request.json()
            query = data.get('query')
            max_results = data.get('max_results', 10)

            if not query:
                return web.json_response({
                    'success': False,
                    'error': 'Query is required',
                    'code': 400
                }, status=400)

            self.request_count += 1

            # Check cache first
            cache_key = f"search:{query}:{max_results}"
            if cache_key in self.cache:
                cache_entry = self.cache[cache_key]
                if datetime.now() - cache_entry['timestamp'] < self.cache_ttl:
                    self.logger.info(f"Cache hit for search: {query}")
                    return web.json_response({
                        'success': True,
                        'data': cache_entry['data']
                    })

            # Perform search using yt-dlp
            loop = asyncio.get_event_loop()
            results = await loop.run_in_executor(
                self.executor,
                self._search_sync,
                query,
                max_results
            )

            if results is not None:
                # Cache the result
                self.cache[cache_key] = {
                    'data': results,
                    'timestamp': datetime.now()
                }

                return web.json_response({
                    'success': True,
                    'data': results
                })
            else:
                self.error_count += 1
                return web.json_response({
                    'success': False,
                    'error': 'Search failed',
                    'code': 500
                }, status=500)

        except Exception as e:
            self.error_count += 1
            self.logger.error(f"Error searching: {str(e)}")
            return web.json_response({
                'success': False,
                'error': str(e),
                'code': 500
            }, status=500)

    def _search_sync(self, query: str, max_results: int) -> Optional[Dict]:
        """Synchronous search implementation"""
        try:
            search_query = f"ytsearch{max_results}:{query}"

            opts = self.ytdl_opts.copy()
            opts['quiet'] = True

            with yt_dlp.YoutubeDL(opts) as ydl:
                search_results = ydl.extract_info(search_query, download=False)

                if not search_results or 'entries' not in search_results:
                    return {
                        'videos': [],
                        'total_count': 0,
                        'query': query
                    }

                videos = []
                for entry in search_results['entries'][:max_results]:
                    if entry:
                        video_info = {
                            'id': entry.get('id', ''),
                            'title': entry.get('title', ''),
                            'description': entry.get('description', ''),
                            'duration': entry.get('duration'),
                            'webpage_url': entry.get('webpage_url', ''),
                            'thumbnail': self._get_best_thumbnail(entry.get('thumbnails', [])),
                            'uploader': entry.get('uploader', ''),
                            'upload_date': entry.get('upload_date', ''),
                            'view_count': entry.get('view_count'),
                            'extractor': entry.get('extractor', ''),
                            'extractor_key': entry.get('extractor_key', ''),
                            'available': True,
                            'live_status': entry.get('live_status'),
                            'formats': self._clean_formats(entry.get('formats', [])),
                            'thumbnails': self._clean_thumbnails(entry.get('thumbnails', []))
                        }
                        videos.append(video_info)

                return {
                    'videos': videos,
                    'total_count': len(videos),
                    'query': query
                }

        except Exception as e:
            self.logger.error(f"Search error for '{query}': {str(e)}")
            return None

    def _get_best_thumbnail(self, thumbnails: List[Dict]) -> str:
        """Get the best quality thumbnail URL"""
        if not thumbnails:
            return ""

        # Sort by preference: width desc, then height desc
        sorted_thumbs = sorted(
            thumbnails,
            key=lambda x: (x.get('width', 0), x.get('height', 0)),
            reverse=True
        )

        return sorted_thumbs[0].get('url', '') if sorted_thumbs else ""

    def _clean_formats(self, formats: List[Dict]) -> List[Dict]:
        """Clean and filter format information"""
        clean_formats = []
        for fmt in formats:
            clean_format = {
                'format_id': fmt.get('format_id', ''),
                'url': fmt.get('url', ''),
                'ext': fmt.get('ext', ''),
                'format': fmt.get('format', ''),
                'protocol': fmt.get('protocol'),
                'vcodec': fmt.get('vcodec'),
                'acodec': fmt.get('acodec'),
                'width': fmt.get('width'),
                'height': fmt.get('height'),
                'fps': fmt.get('fps'),
                'tbr': fmt.get('tbr'),
                'vbr': fmt.get('vbr'),
                'abr': fmt.get('abr'),
                'asr': fmt.get('asr'),
                'filesize': fmt.get('filesize'),
                'quality': fmt.get('quality'),
                'language': fmt.get('language'),
                'preference': fmt.get('preference'),
            }
            clean_formats.append(clean_format)

        return clean_formats

    def _clean_thumbnails(self, thumbnails: List[Dict]) -> List[Dict]:
        """Clean thumbnail information"""
        clean_thumbnails = []
        for thumb in thumbnails:
            clean_thumb = {
                'id': thumb.get('id'),
                'url': thumb.get('url', ''),
                'width': thumb.get('width'),
                'height': thumb.get('height'),
                'resolution': thumb.get('resolution'),
            }
            clean_thumbnails.append(clean_thumb)

        return clean_thumbnails

    async def clear_cache(self, request):
        """Clear the service cache"""
        try:
            self.cache.clear()
            return web.json_response({
                'success': True,
                'data': {'message': 'Cache cleared successfully'}
            })
        except Exception as e:
            return web.json_response({
                'success': False,
                'error': str(e),
                'code': 500
            }, status=500)

    def cleanup_cache(self):
        """Remove expired entries from cache"""
        now = datetime.now()
        expired_keys = []

        for key, entry in self.cache.items():
            if now - entry['timestamp'] > self.cache_ttl:
                expired_keys.append(key)

        for key in expired_keys:
            del self.cache[key]

        if expired_keys:
            self.logger.info(f"Cleaned up {len(expired_keys)} expired cache entries")


async def create_app(config: Dict[str, Any]) -> web.Application:
    """Create the aiohttp application"""
    service = YTDLPService(config)

    app = web.Application()

    # Add routes
    app.router.add_get('/health', service.health_check)
    app.router.add_post('/extract', service.extract_info)
    app.router.add_post('/search', service.search)
    app.router.add_post('/cache/clear', service.clear_cache)

    # Store service instance for cleanup
    app['service'] = service

    return app


async def cleanup_task(app: web.Application):
    """Periodic cleanup task"""
    service = app['service']

    while True:
        await asyncio.sleep(300)  # Run every 5 minutes
        try:
            service.cleanup_cache()
        except Exception as e:
            service.logger.error(f"Cache cleanup error: {str(e)}")


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='yt-dlp HTTP Service')
    parser.add_argument('--config', type=str, help='Configuration file path')
    parser.add_argument('--host', type=str, default='localhost', help='Host to bind to')
    parser.add_argument('--port', type=int, default=8080, help='Port to bind to')
    parser.add_argument('--workers', type=int, default=4, help='Number of worker threads')

    args = parser.parse_args()

    # Load configuration
    config = {
        'host': args.host,
        'port': args.port,
        'max_workers': args.workers,
        'format': 'bestaudio/best',
        'audio_format': 'opus',
        'audio_quality': '128',
        'cache_dir': '/tmp/ytdlp-cache',
        'cache_ttl_hours': 24,
    }

    if args.config and os.path.exists(args.config):
        with open(args.config, 'r') as f:
            file_config = json.load(f)
            config.update(file_config)

    async def init():
        app = await create_app(config)

        # Start cleanup task
        asyncio.create_task(cleanup_task(app))

        return app

    # Run the service
    web.run_app(
        init(),
        host=config['host'],
        port=config['port'],
        access_log_format='%a %t "%r" %s %b "%{Referer}i" "%{User-Agent}i" %Tf'
    )


if __name__ == '__main__':
    main()