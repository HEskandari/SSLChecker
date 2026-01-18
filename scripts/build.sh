#!/bin/bash
# SSL Certificate Monitor Build Script
# Builds the application and copies binaries to dist folder

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="ssl-cert-monitor"
VERSION=$(git describe --tags 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
DIST_DIR="dist"
BUILD_DIR="build"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required commands
check_requirements() {
    print_info "Checking requirements..."
    
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ $(echo "$GO_VERSION < 1.21" | bc -l 2>/dev/null) -eq 1 ]]; then
        print_warn "Go version $GO_VERSION detected. Go 1.21 or later is recommended."
    fi
    
    print_info "Go version: $(go version)"
}

# Clean previous builds
clean() {
    print_info "Cleaning previous builds..."
    rm -rf "$BUILD_DIR"
    rm -rf "$DIST_DIR"
    rm -f "$PROJECT_NAME"
    rm -f test-state.json
}

# Build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local suffix=$3
    
    print_info "Building for $os/$arch..."
    
    local output_name="$PROJECT_NAME"
    if [[ "$os" == "windows" ]]; then
        output_name="$output_name.exe"
    fi
    
    GOOS="$os" GOARCH="$arch" go build \
        -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
        -o "$BUILD_DIR/$PROJECT_NAME-$os-$arch$suffix" \
        ./cmd/ssl-cert-monitor
    
    if [[ $? -eq 0 ]]; then
        print_info "Successfully built for $os/$arch"
    else
        print_error "Failed to build for $os/$arch"
        exit 1
    fi
}

# Build all platforms
build_all() {
    print_info "Building for all platforms..."
    
    mkdir -p "$BUILD_DIR"
    
    # Linux
    build_platform "linux" "amd64" ""
    build_platform "linux" "arm64" ""
    
    # macOS
    build_platform "darwin" "amd64" ""
    build_platform "darwin" "arm64" ""
    
    # Windows
    build_platform "windows" "amd64" ".exe"
    
    print_info "All builds completed successfully"
}

# Create distribution packages
create_dist() {
    print_info "Creating distribution packages..."
    
    mkdir -p "$DIST_DIR"
    
    # Create packages for each platform
    for platform in "$BUILD_DIR"/*; do
        if [[ -f "$platform" ]]; then
            local filename=$(basename "$platform")
            local os_arch=$(echo "$filename" | sed "s/$PROJECT_NAME-//" | sed 's/\.exe$//')
            local os=$(echo "$os_arch" | cut -d'-' -f1)
            local arch=$(echo "$os_arch" | cut -d'-' -f2)
            
            print_info "Creating package for $os/$arch..."
            
            # Create temporary directory for this platform
            local temp_dir="$DIST_DIR/temp-$os-$arch"
            mkdir -p "$temp_dir"
            
            # Copy binary
            cp "$platform" "$temp_dir/$PROJECT_NAME"
            if [[ "$os" == "windows" ]]; then
                mv "$temp_dir/$PROJECT_NAME" "$temp_dir/$PROJECT_NAME.exe"
            fi
            
            # Copy configuration and documentation
            cp config.example.yaml "$temp_dir/"
            cp README.md "$temp_dir/"
            cp -r deployment "$temp_dir/" 2>/dev/null || true
            
            # Create archive
            if [[ "$os" == "windows" ]]; then
                (cd "$temp_dir" && zip -r "../$PROJECT_NAME-$os-$arch.zip" .)
            else
                (cd "$temp_dir" && tar -czf "../$PROJECT_NAME-$os-$arch.tar.gz" .)
            fi
            
            # Clean up
            rm -rf "$temp_dir"
        fi
    done
    
    # Also create a simple binary-only distribution
    print_info "Creating binary-only distribution..."
    mkdir -p "$DIST_DIR/binaries"
    cp "$BUILD_DIR"/* "$DIST_DIR/binaries/"
    
    print_info "Distribution packages created in $DIST_DIR/"
}

# Build for current platform only
build_current() {
    print_info "Building for current platform..."
    
    go build \
        -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
        -o "$PROJECT_NAME" ./cmd/ssl-cert-monitor
    
    if [[ $? -eq 0 ]]; then
        print_info "Successfully built $PROJECT_NAME"
        print_info "Binary: ./$PROJECT_NAME"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Show usage
usage() {
    echo "SSL Certificate Monitor Build Script"
    echo ""
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  all          Build for all platforms and create distribution packages (default)"
    echo "  current      Build only for current platform"
    echo "  clean        Clean build artifacts"
    echo "  help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 all       # Build for all platforms and create dist packages"
    echo "  $0 current   # Build only for current platform"
    echo "  $0 clean     # Clean build artifacts"
}

# Main script
main() {
    local command=${1:-all}
    
    case "$command" in
        all)
            check_requirements
            clean
            build_all
            create_dist
            ;;
        current)
            check_requirements
            clean
            build_current
            ;;
        clean)
            clean
            print_info "Clean completed"
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            print_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"