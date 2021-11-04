var searchjs = {
    data: {
        scope: $("#searchPage"),
        queryString: basejs.getQueryString(),
        loadMorePars: {
            //设置essays加载更多
            offOn: false,
            page: 1

        }
    },
    init: function () {
        var _this = this;
 
        _this.bindSearchResult();
        _this.bindRecommendList();

        $('.go-to-top', _this.data.scope).goToTop();
    },
    template: {
        searchResultRow: ("<div class='search-result-warp'>" +
            "<div class='result-score'><div class='essay-score'>#{start_at}</div><div class='essay-score-head-count'>#{lan}</div></div>" +

            "<div class='result-entity'>" +

            "<div class='result-info'>" +

            "<div class='result-header'><a href='#{essayDetailPage}' class='essay-title'>#{title}</a></div>" +
            "<div class='result-header'><a href='#{essayDetailPage}' class='essay-title'>#{subtitle}</a></div>" +
            "<div class='result-content'><a href='#{essayDetailPage}' class='essay-content'>#{texts}</a></div>" +
           
            "<div class='result-footer'><span class='essay-nickname'><a>#{lan}</a></span> <span class='essay-creationtime'>#{create_time}</span></div>" +

            "</div>" +

            "</div >" +
            "</div >")
    }, 
    bindSearchResult: function () {
        var _this = this;

        var $loadMore = $(".load-more>span", _this.data.scope);
        $loadMore.text("加载中...");

        var keyword = null;
        if (_this.data.queryString && _this.data.queryString.keyword && _this.data.queryString.keyword.length > 0) {
            keyword = decodeURI(_this.data.queryString.keyword);
            $("#searchBox", _this.data.scope).val(keyword);
        }
        var httpPars = {
            url: basejs.requestDomain + "/home/search",
            type: "GET",
            data: { search_word: keyword, page_count: 10 },
            success: function (resultDto) {
                //设置essays加载更多
                if (resultDto.code!=200) {
                    return;
                }

                _this.bindResultInfo(resultDto.data.search_hits);
                //图片懒加载
                $imageLazy = $(".search-result-warp img.lazy", _this.data.scope);
                basejs.lazyInof($imageLazy);
                $imageLazy.removeClass("lazy");

                if (resultDto.data.hasNextPage) {
                    _this.data.loadMorePars.offOn = true;
                    _this.data.loadMorePars.page++;
                    $loadMore.text("加载更多");
                }
                else {
                    _this.data.loadMorePars.offOn = false;
                    $loadMore.text("已经是底部");
                }
            },
            error: function () {
                _this.data.loadMorePars.offOn = true;

                $(".search-result-left", _this.data.scope).empty();
            }
        };

        var essaysHttpHelper = new httpHelper(httpPars);
        essaysHttpHelper.send();

        $(".btn-search", _this.data.scope).click(function () {

            _this.data.loadMorePars.offOn = false;
            _this.data.loadMorePars.page = 1;
            httpPars.data.keyword = $("#searchBox", _this.data.scope).val();
            httpPars.data.pageIndex = _this.data.loadMorePars.page;

            $loadMore.text("加载中...");
            essaysHttpHelper = new httpHelper(httpPars);
            essaysHttpHelper.send();
        });

        $("body").keydown(function () {
            if (event.keyCode == "13") {//keyCode=13是回车键；数字不同代表监听的按键不同
                $(".btn-search", _this.data.scope).click();
            }
        });


        $loadMore.loadMore(50, function () {
            //这里用 [ off_on ] 来控制是否加载 （这样就解决了 当上页的条件满足时，一下子加载多次的问题啦）
            if (_this.data.loadMorePars.offOn) {
                _this.data.loadMorePars.offOn = false;

                httpPars.data.keyword = $("#searchBox", _this.data.scope).val();
                httpPars.data.pageIndex = _this.data.loadMorePars.page;
                $loadMore.text("加载中...");
                essaysHttpHelper = new httpHelper(httpPars);
                essaysHttpHelper.send();
            }
        });

    },
    bindResultInfo: function (data) {

        var _this = this;
        var titleTagArr = [];
        var resultHtml = "";

        if (data) {
            var resultRowHtml = "";

            for (var index in data) {
                var searchHitDto = data[index];
                var essayDetailPage = "/" + searchHitDto.pageUrl;
                var defaultPicturePath = basejs.cdnDomain +"/image/default-picture_100x100.jpg";
                var pictureCropPath = "";
                switch (searchHitDto.coverMediaType) {
                    case "picture": pictureCropPath = basejs.cdnDomain + "/" + searchHitDto.coverPath + "_100x100." + searchHitDto.coverExtension; break;
                    case "video": pictureCropPath = basejs.cdnDomain + "/" + searchHitDto.coverPath + "_100x100.jpg"; break;
                }


                var tagSpan = "";
                if (searchHitDto.tagList && searchHitDto.tagList.length > 0) {
                    tagSpan += "<span class='essay-tag' title='" + searchHitDto.tagList[0].tagName + "'>" + searchHitDto.tagList[0].tagName + "</span>";//(topMediaDto.tagList[0].tagName.length > 4 ? topMediaDto.tagList[0].tagName.substr(0, 3) + "..." : topMediaDto.tagList[0].tagName);
                    titleTagArr.push(searchHitDto.tagList[0].tagName);
                }
                var categorySpan = "<span class='essay-category' title='" + searchHitDto.category + "'>" + searchHitDto.category + "</span>";

                resultRowHtml += _this.template.searchResultRow.format({
                    essayDetailPage: essayDetailPage,
                    defaultPicturePath: defaultPicturePath,
                    pictureCropPath: pictureCropPath,

                    title: searchHitDto.title,
                    subContent: searchHitDto.subContent,

                    score: searchHitDto.score,
                    likeNum:(searchHitDto.likeNum>0?searchHitDto.likeNum + "喜欢":""),
                    creatorNickName: searchHitDto.creatorNickName,

                    browseNum: basejs.getNumberDiff(searchHitDto.browseNum),
                    tagSpan: tagSpan,
                    categorySpan: categorySpan,

                    creationTime: basejs.getDateDiff(basejs.getDateTimeStamp(searchHitDto.creationTime))

                });



                resultHtml += resultRowHtml;
                resultRowHtml = "";

            }

            if (_this.data.loadMorePars.page == 1) {
                $(".search-result-left", _this.data.scope).html(resultHtml);
            } else {
                $(".search-result-left", _this.data.scope).append(resultHtml);
            }
        }


    },
    bindRecommendList: function () {
        var _this=this;
        var httpPars = {
            url: basejs.requestDomain + "/home/essays",
            type: "GET",
            data: { keyword: "", pageIndex: 1, pageSize: 10, orderBy: "choiceness" },
            success: function (resultDto) {

                if (!resultDto.result) {
                    return;
                }
                if ((!resultDto.data)||(!resultDto.data.essayList)||resultDto.data.essayList.length<=0) {
                    return;
                }
                var data=resultDto.data.essayList;
                var essayRecommendAObj = $(".essay-recommend-a", _this.data.scope);

                for (var index in data) {
                    
                    var topMediaDto = data[index];
                     
                    var essayDetailPage = "/" + topMediaDto.pageUrl;
                    essayRecommendAObj.append("<a href='" + essayDetailPage + "' title='"+topMediaDto.title+"'><span class='recommend-list-number'>" + (parseInt(index)+1)+"</span>"+topMediaDto.title + "</a>");
                }
            },
            error: function () {
                _this.data.loadMorePars.offOn = true;

                $(".search-result-left", _this.data.scope).empty();
            }
        };

        var essaysHttpHelper = new httpHelper(httpPars);
        essaysHttpHelper.send();
    }
};

$(function () {
    //菜单
    topMenu.bindMenu();

    searchjs.init();
});