const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
    app.use(
        '/api',
        createProxyMiddleware({
            target: 'http://backend.default:5001/',
            changeOrigin: true,
            pathRewrite: {
                '^/api/': '' // remove base path
            },
        })
    );
};