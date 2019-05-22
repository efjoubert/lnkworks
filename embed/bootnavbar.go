package embed

import (
	"io"
	"strings"
)

const bootnavbarcss = `.dropdown-menu {
    margin-top: 0;
}
.dropdown-menu .dropdown-toggle::after {
    vertical-align: middle;
    border-left: 4px solid;
    border-bottom: 4px solid transparent;
    border-top: 4px solid transparent;
}
.dropdown-menu .dropdown .dropdown-menu {
    left: 100%;
    top: 0%;
    margin:0 20px;
    border-width: 0;
}

.dropdown-menu > li a:hover,
.dropdown-menu > li.show {
	background: #007bff;
	color: white;
}
.dropdown-menu > li.show > a{
	color: white;
}

@media (min-width: 768px) {
    .dropdown-menu .dropdown .dropdown-menu {
        margin:0;
        border-width: 1px;
    }
}
`

const bootnavbarjs = `(function($) {
    var defaults={
        sm : 540,
        md : 720,
        lg : 960,
        xl : 1140,
        navbar_expand: 'lg'
    };
    $.fn.bootnavbar = function() {

        var screen_width = $(document).width();

        if(screen_width >= defaults.lg){
            $(this).find('.dropdown').hover(function() {
                $(this).addClass('show');
                $(this).find('.dropdown-menu').first().addClass('show').addClass('animated fadeIn').one('animationend oAnimationEnd mozAnimationEnd webkitAnimationEnd', function () {
                    $(this).removeClass('animated fadeIn');
                });
            }, function() {
                $(this).removeClass('show');
                $(this).find('.dropdown-menu').first().removeClass('show');
            });
        }

        $('.dropdown-menu a.dropdown-toggle').on('click', function(e) {
          if (!$(this).next().hasClass('show')) {
            $(this).parents('.dropdown-menu').first().find('.show').removeClass("show");
          }
          var $subMenu = $(this).next(".dropdown-menu");
          $subMenu.toggleClass('show');

          $(this).parents('li.nav-item.dropdown.show').on('hidden.bs.dropdown', function(e) {
            $('.dropdown-submenu .show').removeClass("show");
          });

          return false;
        });
    };
})(jQuery);
`

func BootNavbarCSS() io.Reader {
	return strings.NewReader(bootnavbarcss)
}

func BootNavbarJS() io.Reader {
	return strings.NewReader(bootnavbarjs)
}
