const PORT = process.env['SERVER_PORT'] || '8081'; // 8082 for Redis

module.exports = {
  '/api': {
    target: `http://localhost:${PORT}`,
    secure: false,
    changeOrigin: true,
    logLevel: 'warn',
  },
};
