
// 表格弹窗
$(document).on('click','[data-func="pop-wrap"]',function(e){
    var $this = $(this);
    var $content = $this.find('[data-func="pop-content"]');
    if(!$this.hasClass('active')){       
        $('[data-func="pop-wrap"]').removeClass('active');
        $('[data-func="pop-content"]').hide();
        $this.addClass('active');
        $content.show();
    } else {
    	if(!$content.hasClass('active')){   		
        	$content.hide();
        	$this.removeClass('active');
    	}       
    }
    e.stopPropagation();
});

$(document).on('click',function(){
    $('[data-func="pop-wrap"]').removeClass('active');
    $('[data-func="pop-content"]').hide();
});

$(document).on('click','.time-picker-section input',function(e){
	
	e.stopPropagation();
})

$(document).ready(function () {
    //页面回车触发查询
    $("div.search input[type=text].sear-inp").keydown(function(e){
        var input = $(this);
        if (e.which == 13){     
            input.siblings(".sear-btn").click();
        }
    });
});