#!/bin/sh
cat <<EOF > /usr/share/nginx/html/env-config.js
window._env_ = {
  VITE_API_BASE_URL: "${VITE_API_BASE_URL:-http://192.168.9.250}",
};
EOF
