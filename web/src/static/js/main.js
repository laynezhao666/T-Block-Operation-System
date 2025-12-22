/**
 * 
 *
 * 框架js
 *
 * */

//收缩
$(document).on("click", ".left-aside-footer .fa-angle-left", function(e){
    $(".left-aside").addClass("left-aside-small");
    $(".right").addClass("expanded");
    $(".left-header-title").addClass("display-none").addClass("");
    $(".left-aside-header>a>img").attr("src", "/static/images/logo/logo_small.png");
    $(this).removeClass("fa-angle-left").addClass("fa-angle-right");
    $(".left-aside-footer").removeClass("flex-justify-end").addClass("flex-justify-middle");
    window.dispatchEvent(new Event("resize"));  //抛出resize事件，供有需要的地方响应，比如echart大小调整
    // window.sessionStorage.setItem('hideLeftNav', true);
    setHideLeftNav('true');
});

//还原
$(document).on("click", ".left-aside-footer .fa-angle-right", function(e){
    $(".left-aside").removeClass("left-aside-small"); 
    $(".right").removeClass("expanded");
    $(".left-aside-header>a>img").attr("src", "/static/images/logo/logo.png");
    $(".left-header-title").removeClass("display-none");       
    $(this).addClass("fa-angle-left").removeClass("fa-angle-right");
    $(".left-aside-footer").removeClass("flex-justify-middle").addClass("flex-justify-end");
    window.dispatchEvent(new Event("resize"));
    // window.sessionStorage.setItem('hideLeftNav', false);
    setHideLeftNav('false');
});

//显示右侧弹窗
$(document).on("click", ".btn-popup-panel", function(e){
    var btn = $(this);
    var div = btn.attr('data-panel-id');
    $('body').css('overflow-y', 'hidden');
    if ( div != ""){
        $("#" + div).fadeIn(300);
        $("#" + div).find('.tnbl-panel').animate({left: 'auto', right: '0'},300,'linear');
        resizePanelBody();
    }  
    e.stopPropagation();
});

//关闭右侧弹窗，按钮，小图标，遮罩
$(document).on("click", ".btn-close-popup, i.fa-close, .tnbl-panel-popup-right .tnbl-panel-mask", function(e){
//  $(this).parents(".tnbl-panel-popup-right").hide();     
    
    var w ='-' + $(this).parents(".tnbl-panel-popup-right").find(".tnbl-panel").width();

    $(this).parents(".tnbl-panel-popup-right").find(".tnbl-panel").animate({right: w},300,'linear');
    $(this).parents(".tnbl-panel-popup-right").fadeOut(300).trigger('popRightClose');
    $('body').css('overflow-y', 'auto');
});

$(document).on("click", function(e){
    $(this).trigger('popRightClose');
});

//调整可见区域大小，显示滚动条
$(window).on("resize", function(){
    if ( $(".tnbl-panel-popup-right").is(":visible")){
        resizePanelBody();
    }
});

// function resizePanelBody(){
//     var h = $(document).height() -  $(".tnbl-panel-top").height() - $(".tnbl-panel-footer").height() -48;//48为block的margin
//     $(".tnbl-panel-body").css("max-height", h + "px");
// }

function resizePanelBody(){
    var h = $('body').height() -  $(".tnbl-panel-top").height() - $(".tnbl-panel-footer").height();
    $(".tnbl-panel-body").css("max-height", h + "px");
}

//ul展开和收起
$(document).on("click", ".tnbl-ul>li", function(e){
	if($(this).hasClass('active')){
		$(this).removeClass('active');
		$(this).find('.detail').slideUp();
	}else{
		$('.tnbl-ul>li').removeClass('active');
		$(this).addClass('active');
		$('.tnbl-ul>li .detail').slideUp();
		$(this).find('.detail').slideDown();
	}
	e.stopPropagation();
});

//搜索框
$(document).on("click", ".comp-search", function(e){
    var target=e.target;
    if($(target).hasClass('btn-search')){
		$(this).find('.search-inp>input').focus();
	}else if($(target).hasClass('btn-more')){
		$(this).find('.more-select').slideDown(300);
	}
	e.stopPropagation();
});

// 动画效果搜索框
// $(document).on("click", ".comp-search", function(e){
// 	var target=e.target;
// 	if($(target).hasClass('btn-search')){
// 		$(this).find('.search-inp').slideDown(300);
// 		$(this).find('.search-inp>input').focus();
// 	}else if($(target).hasClass('btn-more')){
// 		$(this).find('.more-select').slideDown(300);
// 	}
// 	e.stopPropagation();
// });

// $(document).on("blur", ".comp-search>.search-inp>input", function(e){
// 	$(this).parents('.search-inp').slideUp(300);
// });

$(document).on("click", function(e){
	$('.comp-search>.more-select').slideUp(300);
    $('.right > .menu > .topbar > .wrap-icons > .user-layout > .logout').addClass('hidden');
    $('.right > .menu > .topbar > .wrap-icons > .user-layout > i').addClass('user-layout-idown').removeClass('user-layout-iip');
    $('.right > .menu > .top-icons > .user-layout > .logout').addClass('hidden');
    $('.right > .menu > .topbar > .wrap-icons > .user-layout > i').addClass('user-layout-idown').removeClass('user-layout-iip');
});

$("#btn-close-tips").on("click", function(){
    $(".tnbl-message").closeTips();
});

$("#btn-show-tips").on("click", function(){
    $(".tnbl-message").showTips('<i class="fa fa-frown-o f16 mr20"></i>导入失败，请重新导入');
});

$("#btn-close").on("click", function(){
	console.log("click close button");
});

//showTips    
$.fn.extend({       
    showTips: function(html){
        var tmpl = ['<div class="tnbl-tips">',
            '<span class="text" id="tnbl-tips-text">',html,'</span>',
            '</div>'].join('');            
        $(tmpl).prependTo($(this)).fadeIn(300);
    },
    closeTips: function(){
        $(".tnbl-tips").fadeOut(300, function(){
            $(this).remove();
        });
    }
}); 

var TNBL = TNBL || {};

$(document).ready(function(){
    $("#sel-module-chosen").chosen({
        allow_single_deselect: true,  //是否允许取消选
        disable_search:false,   //设置下拉多选框可见
        search_contains:true,
        no_results_text:"输入信息不正确",
        width: "208px"
    });
    //模组选择
    $("#sel-module-chosen").on("change", function(){
        var moduleName = $("#sel-module-chosen option:selected").text(), moduleId = $(this).val();
        
        if (moduleName == ""){
            return false;
        }
        
        TNBL.ajax({
            url: "/web/core./switchModule" ,
            type: 'POST',
            data: {
                moduleName: moduleName, 
                moduleId: moduleId
            },
            retOk:function(msg, data){
                //如果有新的地址，做跳转
                if (TNBL.afterSwitchModuleUrl){
					if (TNBL.afterSwitchModuleUrl !== '0'){					
	                    location.href = TNBL.afterSwitchModuleUrl;
					}
                }else{
                    window.location.reload();
                }
            },
            retError:function(msg , ret){
                TNBL.showErr(msg);
            }
        });
    });
    //左侧导航是否隐藏,根据cookie中 hiddeLeftNav设置。
    // var hideLeftNav = getCookie('tnebula_hideLeftNav');
    // if(hideLeftNav !== 'true'){
    //     setHideLeftNav('false');
    // }
});

$(document).on('click', '.right > .menu > .top-icons > .user-layout', function (e) {
    $(this).children('i').toggleClass('user-layout-iup').toggleClass('user-layout-idown');
    // if( !$(this).children('i').hasClass('fa-angle-down') )
    // {
    //     $('.right > .menu > .top-icons > .user-layout > .logout').removeClass('hidden');
    // }
    // else
    // {
    //     $('.right > .menu > .top-icons > .user-layout > .logout').addClass('hidden');
    // }
    $('.right > .menu > .top-icons > .user-layout > .logout').toggleClass('hidden');
    e.stopPropagation();
});

$(document).on('scroll', function(e) {
    var scrollTop = $(document).scrollTop();
    if(scrollTop > 72) {
        $('.topbar').hide()
        $('.menu').css('height', '48px');
        $('.top-icons').css('margin-top', '7px');
    }
    if(scrollTop === 0) {       
        $('.topbar').show()
        $('.menu').css('height', '120px');
        $('.top-icons').css('margin-top', '20px');
    }
})

function getCookie(name){
    var arr = document.cookie.match(new RegExp("(^| )"+name+"=([^;]*)(;|$)"));
    if(arr != null) return unescape(arr[2]);
    return null;
}
function setHideLeftNav(value) {
    var domainArr = document.domain.split('.');
    if(domainArr.length < 2) return;
    var len = domainArr.length;
    var domain = domainArr[len-2] + '.' + domainArr[len-1];
    var expireArr = getCookie('tnebula_expire').split('+');
    var dateArr = expireArr[0].split('-');
    var timeArr = expireArr[1].split(':');
    var expire = new Date(dateArr[0],dateArr[1]-1,dateArr[2],timeArr[0],timeArr[1],timeArr[2]);
    document.cookie = ' tnebula_hideLeftNav=' + value + '; domain=' + domain + '; path=/; expires=' + expire;
}