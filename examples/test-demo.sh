#!/bin/bash

# 构建项目
echo "正在构建项目..."
go build -o git-watcher

if [ $? -ne 0 ]; then
    echo "构建失败"
    exit 1
fi

echo "构建成功！"

# 创建测试用的Git仓库
TEST_DIR="./test-repos"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# 创建第一个测试仓库
echo "创建测试仓库1..."
mkdir repo1
cd repo1
git init
git config user.name "测试用户1"
git config user.email "test1@example.com"

# 创建一些提交
echo "console.log('hello');" > app.js
git add app.js
git commit -m "初始提交"

echo "console.log('world');" >> app.js
git add app.js
git commit -m "添加功能"

# 模拟深夜提交
GIT_AUTHOR_DATE="2024-01-15 23:30:00" GIT_COMMITTER_DATE="2024-01-15 23:30:00" git commit --allow-empty -m "深夜提交1"

cd ..

# 创建第二个测试仓库
echo "创建测试仓库2..."
mkdir repo2
cd repo2
git init
git config user.name "测试用户2"
git config user.email "test2@example.com"

echo "print('hello')" > main.py
git add main.py
git commit -m "Python初始提交"

echo "print('world')" >> main.py
git add main.py
git commit -m "添加输出"

cd ../..

echo "测试仓库创建完成！"
echo ""
echo "运行Git Watcher进行测试:"
echo "1. JSON格式输出:"
./git-watcher -p "$TEST_DIR" -o json

echo ""
echo "2. 文本格式输出:"
./git-watcher -p "$TEST_DIR" -o text

echo ""
echo "测试完成！"