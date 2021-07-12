const i_path = require('path');

const ZOEKT_APP_DATA_BASE_DIR = (
   process.env.ZOEKT_APP_DATA_BASE_DIR?
   i_path.resolve(process.env.ZOEKT_APP_DATA_BASE_DIR):null
);
if (!ZOEKT_APP_DATA_BASE_DIR) {
   console.warn('[!] ZOEKT_APP_DATA_BASE_DIR is empty: app persistence is disabled');
}

const api = {
   config: {
      baseDir: ZOEKT_APP_DATA_BASE_DIR
   },
};

module.exports = api;
