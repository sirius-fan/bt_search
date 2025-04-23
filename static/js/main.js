document.addEventListener('DOMContentLoaded', function() {
    // 定义全局变量
    let currentPage = 1;
    let currentQuery = '';
    let currentSort = '';
    let currentOrder = 'desc';
    
    // 获取DOM元素
    const searchForm = document.getElementById('search-form');
    const searchInput = document.getElementById('search-input');
    const resultsContainer = document.getElementById('results-container');
    const paginationContainer = document.getElementById('pagination');
    const sortOptions = document.getElementById('sort-options');
    const resultStats = document.getElementById('result-stats');

    // 绑定搜索表单提交事件
    searchForm.addEventListener('submit', function(e) {
        e.preventDefault();
        currentQuery = searchInput.value.trim();
        currentPage = 1;
        searchTorrents();
    });

    // 绑定排序选项点击事件
    document.querySelectorAll('.sort-option').forEach(option => {
        option.addEventListener('click', function(e) {
            e.preventDefault();
            currentSort = this.dataset.sort;
            currentOrder = this.dataset.order;
            searchTorrents();
        });
    });

    // 检查URL参数，如果有任何参数则使用它们
    const urlParams = new URLSearchParams(window.location.search);
    
    // 先检查是否有页码参数，这很重要
    if (urlParams.has('page')) {
        currentPage = parseInt(urlParams.get('page')) || 1;
    }
    
    // 然后检查其他参数
    if (urlParams.has('q')) {
        currentQuery = urlParams.get('q');
        searchInput.value = currentQuery;
    }
    
    if (urlParams.has('sort')) {
        currentSort = urlParams.get('sort');
    }
    
    if (urlParams.has('order')) {
        currentOrder = urlParams.get('order');
    }
    
    // 无论有没有参数，都执行搜索
    searchTorrents();

    // 执行搜索的函数
    function searchTorrents() {
        // 先确保currentPage是有效数字
        if (!currentPage || isNaN(currentPage) || currentPage < 1) {
            currentPage = 1;
        }
        
        // console.log('搜索中，当前页码:', currentPage); // 调试信息
        
        // 更新URL，方便分享和刷新
        updateURL();
        
        // 显示加载状态
        resultsContainer.innerHTML = '<div class="text-center my-5"><div class="spinner-border text-primary" role="status"><span class="visually-hidden">Loading...</span></div></div>';
        
        // 构建API请求URL - 确保页码参数总是被包含
        let apiUrl = `/api/search?page=${currentPage}`;
        if (currentQuery) {
            apiUrl += `&q=${encodeURIComponent(currentQuery)}`;
        }
        if (currentSort) {
            apiUrl += `&sort=${currentSort}&order=${currentOrder}`;
        }
        
        // console.log('API请求URL:', apiUrl); // 调试信息

        // 发送请求
        fetch(apiUrl)
            .then(response => {
                if (!response.ok) {
                    throw new Error('网络响应错误');
                }
                return response.json();
            })
            .then(data => {
                // 同步API返回的页码，确保前端和后端一致
                if (data.page && !isNaN(data.page)) {
                    currentPage = parseInt(data.page);
                }
                
                displayResults(data);
                // 显示排序选项
                sortOptions.classList.remove('d-none');
                // 显示点击提示
                document.getElementById('tip-container').classList.remove('d-none');
                // 更新结果统计
                resultStats.textContent = `找到 ${data.total} 个结果，第 ${currentPage} 页`;
            })
            .catch(error => {
                resultsContainer.innerHTML = `<div class="alert alert-danger" role="alert">
                    搜索出错: ${error.message}
                </div>`;
            });
    }

    // 显示搜索结果
    function displayResults(data) {
        if (data.results.length === 0) {
            resultsContainer.innerHTML = `<div class="alert alert-info" role="alert">
                未找到相关结果
            </div>`;
            paginationContainer.innerHTML = '';
            return;
        }

        let resultsHTML = '';
        
        // 遍历结果并生成HTML
        data.results.forEach(torrent => {
            const createDate = new Date(torrent.create_date);
            const formattedDate = createDate.toLocaleDateString('zh-CN');
            
            // 格式化文件大小
            const sizeGB = (torrent.total_size / 1073741824).toFixed(2); // 转换为GB
            
            // 确保info_hash不为空
            if (!torrent.info_hash) {
                console.warn('遇到没有info_hash的种子:', torrent);
            }
            
            const infoHash = torrent.info_hash || '';
            
            resultsHTML += `
                <div class="torrent-card" data-info-hash="${infoHash}">
                    <h3 class="torrent-name">${escapeHTML(torrent.name)}</h3>
                    <div class="torrent-info d-flex flex-wrap justify-content-between align-items-center">
                        <div>
                            <span class="badge bg-primary me-2">文件数: ${torrent.file_count}</span>
                            <span class="badge bg-success me-2 torrent-size">${sizeGB} GB</span>
                            <span class="badge bg-secondary">${formattedDate}</span>
                        </div>
                        <div class="mt-2 mt-md-0">
                            <a href="/torrent/${infoHash}" class="btn btn-sm btn-outline-primary detail-btn">查看详情</a>
                            <a href="magnet:?xt=urn:btih:${infoHash}" class="btn btn-sm btn-primary magnet-btn">磁力链接</a>
                        </div>
                    </div>
                </div>
            `;
        });

        resultsContainer.innerHTML = resultsHTML;
        
        // 为每个卡片添加点击事件
        document.querySelectorAll('.torrent-card').forEach(card => {
            card.addEventListener('click', function(e) {
                // 如果点击的是按钮或按钮内的元素，不触发卡片的点击事件
                if (e.target.closest('.detail-btn') || e.target.closest('.magnet-btn')) {
                    return;
                }
                
                // 获取种子的info_hash
                const infoHash = this.getAttribute('data-info-hash');
                // console.log('点击卡片，info_hash:', infoHash); // 调试信息
                
                if (!infoHash) {
                    console.error('未找到info_hash属性');
                    return;
                }
                
                // 添加点击反馈效果
                this.classList.add('card-clicked');
                
                // 短暂延迟以显示点击效果
                setTimeout(() => {
                    // 跳转到详情页
                    window.location.href = `/torrent/${infoHash}`;
                }, 150);
            });
            
            // 添加手型光标，提示可点击
            card.style.cursor = 'pointer';
        });
        
        // 生成分页控件 - 确保使用当前页码，而不是依赖API返回值
        // (如果API返回的页码与当前页码不同，我们在之前已同步)
        generatePagination(currentPage, Math.ceil(data.total / data.per_page));
    }

    // 生成分页控件
    function generatePagination(curPage, totalPages) {
        // 确保参数有效
        curPage = parseInt(curPage) || 1;
        // console.log('生成分页组件，当前页:', curPage, '总页数:', totalPages);
        
        if (totalPages <= 1) {
            paginationContainer.innerHTML = '';
            return;
        }

        let paginationHTML = '<ul class="pagination">';
        
        // 上一页按钮
        if (currentPage > 1) {
            paginationHTML += `<li class="page-item"><a class="page-link" href="#" data-page="${currentPage - 1}">上一页</a></li>`;
        } else {
            paginationHTML += '<li class="page-item disabled"><span class="page-link">上一页</span></li>';
        }
        
        // 页码按钮
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, startPage + 4);
        
        // 第一页
        if (startPage > 1) {
            paginationHTML += `<li class="page-item"><a class="page-link" href="#" data-page="1">1</a></li>`;
            if (startPage > 2) {
                paginationHTML += '<li class="page-item disabled"><span class="page-link">...</span></li>';
            }
        }
        
        // 页码
        for (let i = startPage; i <= endPage; i++) {
            if (i === currentPage) {
                paginationHTML += `<li class="page-item active"><span class="page-link">${i}</span></li>`;
            } else {
                paginationHTML += `<li class="page-item"><a class="page-link" href="#" data-page="${i}">${i}</a></li>`;
            }
        }
        
        // 最后一页
        if (endPage < totalPages) {
            if (endPage < totalPages - 1) {
                paginationHTML += '<li class="page-item disabled"><span class="page-link">...</span></li>';
            }
            paginationHTML += `<li class="page-item"><a class="page-link" href="#" data-page="${totalPages}">${totalPages}</a></li>`;
        }
        
        // 下一页按钮
        if (currentPage < totalPages) {
            paginationHTML += `<li class="page-item"><a class="page-link" href="#" data-page="${currentPage + 1}">下一页</a></li>`;
        } else {
            paginationHTML += '<li class="page-item disabled"><span class="page-link">下一页</span></li>';
        }
        
        paginationHTML += '</ul>';
        
        paginationContainer.innerHTML = paginationHTML;
        
        // 绑定分页点击事件
        paginationContainer.querySelectorAll('.page-link').forEach(link => {
            if (link.dataset.page) {
                link.addEventListener('click', function(e) {
                    e.preventDefault();
                    // 确保转换为数字
                    const newPage = parseInt(this.dataset.page);
                    // 只有当页数真的改变时才执行搜索
                    if (currentPage !== newPage) {
                        currentPage = newPage;
                        // console.log('页码已更新为:', currentPage); // 调试信息
                        searchTorrents();
                        // 滚动到顶部
                        window.scrollTo(0, 0);
                    }
                });
            }
        });
    }

    // 更新URL，方便分享和刷新
    function updateURL() {
        const params = new URLSearchParams();
        if (currentQuery) {
            params.set('q', currentQuery);
        }
        // 总是记录当前页码，不管是否为1
        params.set('page', currentPage);
        
        if (currentSort) {
            params.set('sort', currentSort);
            params.set('order', currentOrder);
        }
        
        const newURL = window.location.pathname + (params.toString() ? '?' + params.toString() : '');
        window.history.pushState({}, '', newURL);
    }

    // HTML转义函数，防止XSS攻击
    function escapeHTML(str) {
        return str
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }
});
