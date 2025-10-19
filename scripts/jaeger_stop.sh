#!/bin/bash

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Остановка Jaeger по PID файлу
stop_by_pid() {
    if [ -f /tmp/jaeger.pid ]; then
        local pid=$(cat /tmp/jaeger.pid)
        if kill -0 $pid 2>/dev/null; then
            log_info "Stopping Jaeger process (PID: $pid)..."
            kill $pid
            sleep 2
            
            # Проверяем, что процесс остановился
            if kill -0 $pid 2>/dev/null; then
                log_warning "Process still running, force killing..."
                kill -9 $pid
            fi
            
            log_success "Jaeger process stopped"
            rm -f /tmp/jaeger.pid
        else
            log_warning "PID file exists but process is not running"
            rm -f /tmp/jaeger.pid
        fi
    else
        log_info "No PID file found"
    fi
}

# Остановка всех процессов Jaeger
stop_all_jaeger() {
    local pids=$(pgrep -f "jaeger" || true)
    if [ -n "$pids" ]; then
        log_info "Found Jaeger processes: $pids"
        log_info "Stopping all Jaeger processes..."
        
        # Сначала пробуем мягкую остановку
        pkill -f "jaeger" || true
        sleep 3
        
        # Проверяем, остались ли процессы
        local remaining_pids=$(pgrep -f "jaeger" || true)
        if [ -n "$remaining_pids" ]; then
            log_warning "Some processes are still running, force killing..."
            pkill -9 -f "jaeger" || true
            sleep 1
        fi
        
        log_success "All Jaeger processes stopped"
    else
        log_info "No Jaeger processes found"
    fi
}

# Проверка портов
check_ports() {
    local ports=(16686 14268 14250 4317 4318)
    local occupied_ports=()
    
    for port in "${ports[@]}"; do
        if lsof -i :$port >/dev/null 2>&1; then
            occupied_ports+=($port)
        fi
    done
    
    if [ ${#occupied_ports[@]} -gt 0 ]; then
        log_warning "The following ports are still in use: ${occupied_ports[*]}"
        log_warning "You may need to stop other services using these ports"
    else
        log_success "All Jaeger ports are free"
    fi
}

# Очистка временных файлов
cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f /tmp/jaeger.pid
    log_success "Cleanup completed"
}

# Основная функция
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}        Jaeger All-in-One Stopper${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo
    
    stop_by_pid
    stop_all_jaeger
    check_ports
    cleanup
    
    log_success "Jaeger has been stopped successfully!"
    echo
}

# Обработка сигналов
trap 'log_warning "Script interrupted"; exit 1' INT TERM

# Запуск основной функции
main "$@"




