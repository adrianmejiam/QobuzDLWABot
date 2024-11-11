import logging
import os
from dotenv.main import dotenv_values
from flask import Flask
from flask import request
from qobuz_dl.core import QobuzDL
from qobuz_dl.utils import get_url_info
from qobuz_dl.downloader import _get_title, _safe_get
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(level=logging.WARN)

app = Flask(__name__)

email = os.getenv('QOBUZ_EMAIL')
password = os.getenv('QOBUZ_PASSWORD')

qobuz = QobuzDL()
qobuz.get_tokens()
qobuz.initialize_client(email, password, qobuz.app_id, qobuz.secrets)

@app.route('/download')
def download():
    url = request.args.get('url')
    try:
        qobuz.handle_url(url)
        return get_file_path(url)
    except Exception as e:
        print(e)
    return f'Failed to get info from url "{url}"'
    
def get_file_path(url):
    url_type, item_id = get_url_info(url)
    is_album = False
    try:
        meta = qobuz.client.get_track_meta(item_id)
    except Exception:
        is_album = True
        try: 
            meta = qobuz.client.get_album_meta(item_id)
        except Exception:
            return None
    
    title = _get_title(meta)
    artist = _safe_get(meta, "performer", "name") if not is_album else _safe_get(meta, "artist", "name")
    
    for root, dirs, files in os.walk(qobuz.directory):
        for dir in dirs:
            if not title in dir or not artist in dir: continue
            return dir
    
    return None

if __name__ == "__main__":
    app.run(port=8000, debug=True)