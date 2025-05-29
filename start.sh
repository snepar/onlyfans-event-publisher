#!/bin/bash

# OnlyFans Event Publisher - Quick Start Script

set -e

echo "🚀 OnlyFans Event Publisher - Quick Start"
echo "========================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo "❌ docker-compose is not installed. Please install it and try again."
    exit 1
fi

echo "✅ Docker is running"

# Build and start services
echo ""
echo "📦 Building application..."
docker-compose build --no-cache onlyfans-publisher

echo ""
echo "🏃 Starting Redpanda cluster and application..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be ready..."

# Wait for Redpanda to be healthy
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker-compose exec -T redpanda-1 rpk cluster health > /dev/null 2>&1; then
        echo "✅ Redpanda cluster is healthy"
        break
    fi
    attempt=$((attempt + 1))
    echo "   Waiting for Redpanda... ($attempt/$max_attempts)"
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Redpanda cluster failed to start properly"
    echo "   Check logs with: docker-compose logs redpanda-1"
    exit 1
fi

# Wait a bit more for the application to start
sleep 5

echo ""
echo "🎉 Setup complete!"
echo ""
echo "📊 Access Points:"
echo "   • Redpanda Console: http://localhost:8080"
echo "   • Application logs: docker-compose logs -f onlyfans-publisher"
echo ""
echo "🔧 Management Commands:"
echo "   • View all logs:    docker-compose logs -f"
echo "   • Stop services:    docker-compose down"
echo "   • Restart:          docker-compose restart"
echo "   • Check status:     docker-compose ps"
echo ""
echo "📈 Topics created:"
docker-compose exec -T redpanda-1 rpk topic list 2>/dev/null || echo "   (Topics will be created automatically)"

echo ""
echo "🔥 The simulation is now running!"
echo "   Watch the logs to see events being published:"
echo "   docker-compose logs -f onlyfans-publisher"
echo ""
echo "   Press Ctrl+C to stop following logs (services will keep running)"
echo "   Use 'docker-compose down' to stop all services"
echo ""

# Follow application logs
echo "📝 Following application logs (Ctrl+C to exit):"
echo "================================================"
docker-compose logs -f onlyfans-publisher