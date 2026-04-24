#!/bin/bash

APP_NAME="ATDReminder"
OUTPUT_DIR="build"
PLATFORMS=("windows/amd64" "linux/amd64" "darwin/arm64" )
# PLATFORMS=("darwin/arm64") # Builld on MacOS
# PLATFORMS=("windows/arm64") # Builld on Windows
# PLATFORMS=("windows/amd64") # Builld on Windows

echo "Cleaning up previous builds..."
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

for PLATFORM in "${PLATFORMS[@]}"; do
    # Tách GOOS và GOARCH
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    
    OUTPUT_NAME="$OUTPUT_DIR/$GOOS-$GOARCH/$APP_NAME-$GOOS-$GOARCH"
    
    # Cấu hình LDFLAGS cơ bản để giảm dung lượng file (-s -w)
    # -s: xóa symbol table, -w: xóa debug info
    CORE_LDFLAGS="-s -w"
    CGO_ENABLED=0
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
        # Thêm flag ẩn console cho Windows
        CURRENT_LDFLAGS="$CORE_LDFLAGS -H=windowsgui"
        # CURRENT_LDF`LAGS="$CORE_LDFLAGS"
    else
        CGO_ENABLED=1
        CURRENT_LDFLAGS="$CORE_LDFLAGS"
    fi

    if [ "$GOOS" == "windows" ]; then
        rsrc -manifest main.manifest -ico main.ico -o main.syso
    fi

    echo "Building for $GOOS/$GOARCH..."
    
    # Sử dụng CGO_ENABLED=0 để đảm bảo tính di động (Static linking)
    echo CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags \"$CURRENT_LDFLAGS\" -o \"$OUTPUT_NAME\"

    env CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags "$CURRENT_LDFLAGS" -o "$OUTPUT_NAME" .

    if [ $? -eq 0 ]; then
        echo "Successfully built: $OUTPUT_NAME"
    else
        echo "Failed to build for $GOOS/$GOARCH"
        exit 1
    fi

    if [ "$GOOS" == "windows" ]; then
        rm main.syso
    fi
done

echo "---------------------------------------"
echo "Build completed! Check the '$OUTPUT_DIR' folder."