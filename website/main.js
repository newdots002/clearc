// ClearC Website JavaScript

document.addEventListener('DOMContentLoaded', function() {
    // Mobile Menu Toggle
    const mobileMenuBtn = document.querySelector('.mobile-menu-btn');
    const mobileNav = document.querySelector('.mobile-nav');
    
    if (mobileMenuBtn && mobileNav) {
        mobileMenuBtn.addEventListener('click', function() {
            mobileNav.classList.toggle('active');
            mobileMenuBtn.classList.toggle('active');
        });
        
        // Close mobile menu when clicking a link
        mobileNav.querySelectorAll('a').forEach(link => {
            link.addEventListener('click', () => {
                mobileNav.classList.remove('active');
                mobileMenuBtn.classList.remove('active');
            });
        });
    }
    
    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            const href = this.getAttribute('href');
            if (href === '#') return;
            
            e.preventDefault();
            const target = document.querySelector(href);
            if (target) {
                const headerHeight = document.querySelector('.header').offsetHeight;
                const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - headerHeight;
                window.scrollTo({
                    top: targetPosition,
                    behavior: 'smooth'
                });
            }
        });
    });
    
    // Header scroll effect
    const header = document.querySelector('.header');
    let lastScroll = 0;
    
    window.addEventListener('scroll', () => {
        const currentScroll = window.pageYOffset;
        
        if (currentScroll > 100) {
            header.style.background = 'rgba(10, 15, 28, 0.95)';
        } else {
            header.style.background = 'rgba(10, 15, 28, 0.8)';
        }
        
        lastScroll = currentScroll;
    });
    
    // Intersection Observer for animations
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate-in');
                observer.unobserve(entry.target);
            }
        });
    }, observerOptions);
    
    // Observe elements for animation
    document.querySelectorAll('.feature-card, .pricing-card, .faq-item').forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(20px)';
        observer.observe(el);
    });
    
    // Add animation styles
    const style = document.createElement('style');
    style.textContent = `
        .animate-in {
            animation: fadeInUp 0.5s ease forwards;
        }
    `;
    document.head.appendChild(style);
    
    // Initialize Image Lightbox
    initLightbox();
});

// Image Lightbox functionality
function initLightbox() {
    // Create lightbox elements
    const lightbox = document.createElement('div');
    lightbox.className = 'lightbox';
    lightbox.innerHTML = `
        <div class="lightbox-overlay"></div>
        <div class="lightbox-content">
            <img class="lightbox-image" src="" alt="">
            <button class="lightbox-close" aria-label="关闭">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
            </button>
            <div class="lightbox-caption"></div>
        </div>
        <button class="lightbox-nav lightbox-prev" aria-label="上一张">
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
        </button>
        <button class="lightbox-nav lightbox-next" aria-label="下一张">
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="9 18 15 12 9 6"></polyline>
            </svg>
        </button>
    `;
    document.body.appendChild(lightbox);
    
    // Add lightbox styles
    const lightboxStyles = document.createElement('style');
    lightboxStyles.textContent = `
        .lightbox {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            z-index: 10000;
            display: flex;
            align-items: center;
            justify-content: center;
            opacity: 0;
            visibility: hidden;
            transition: opacity 0.3s ease, visibility 0.3s ease;
        }
        
        .lightbox.active {
            opacity: 1;
            visibility: visible;
        }
        
        .lightbox-overlay {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.9);
            backdrop-filter: blur(10px);
        }
        
        .lightbox-content {
            position: relative;
            max-width: 90vw;
            max-height: 85vh;
            z-index: 1;
        }
        
        .lightbox-image {
            max-width: 90vw;
            max-height: 80vh;
            object-fit: contain;
            border-radius: 8px;
            box-shadow: 0 25px 80px rgba(0, 0, 0, 0.5);
            transform: scale(0.8);
            opacity: 0;
            transition: transform 0.4s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.3s ease;
        }
        
        .lightbox.active .lightbox-image {
            transform: scale(1);
            opacity: 1;
        }
        
        .lightbox-close {
            position: absolute;
            top: -50px;
            right: 0;
            width: 44px;
            height: 44px;
            background: rgba(255, 255, 255, 0.1);
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 50%;
            color: white;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.2s ease;
        }
        
        .lightbox-close:hover {
            background: rgba(255, 255, 255, 0.2);
            transform: rotate(90deg);
        }
        
        .lightbox-caption {
            text-align: center;
            color: rgba(255, 255, 255, 0.8);
            font-size: 14px;
            margin-top: 16px;
            padding: 0 20px;
        }
        
        .lightbox-nav {
            position: absolute;
            top: 50%;
            transform: translateY(-50%);
            width: 50px;
            height: 50px;
            background: rgba(255, 255, 255, 0.1);
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 50%;
            color: white;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.2s ease;
            z-index: 2;
        }
        
        .lightbox-nav:hover {
            background: rgba(34, 211, 238, 0.3);
            border-color: rgba(34, 211, 238, 0.5);
        }
        
        .lightbox-prev {
            left: 20px;
        }
        
        .lightbox-next {
            right: 20px;
        }
        
        /* Make story card images clickable */
        .story-card-image {
            cursor: zoom-in;
        }
        
        .story-card-image img {
            transition: transform 0.3s ease;
        }
        
        .story-card-image:hover img {
            transform: scale(1.05);
        }
        
        @media (max-width: 768px) {
            .lightbox-nav {
                width: 40px;
                height: 40px;
            }
            
            .lightbox-prev {
                left: 10px;
            }
            
            .lightbox-next {
                right: 10px;
            }
            
            .lightbox-close {
                top: -45px;
                width: 40px;
                height: 40px;
            }
        }
    `;
    document.head.appendChild(lightboxStyles);
    
    // Get elements
    const overlay = lightbox.querySelector('.lightbox-overlay');
    const closeBtn = lightbox.querySelector('.lightbox-close');
    const lightboxImg = lightbox.querySelector('.lightbox-image');
    const caption = lightbox.querySelector('.lightbox-caption');
    const prevBtn = lightbox.querySelector('.lightbox-prev');
    const nextBtn = lightbox.querySelector('.lightbox-next');
    
    // Get all story card images
    let images = [];
    let currentIndex = 0;
    
    function updateImages() {
        images = Array.from(document.querySelectorAll('.story-card-image img'));
    }
    
    function openLightbox(index) {
        updateImages();
        if (images.length === 0) return;
        
        currentIndex = index;
        const img = images[currentIndex];
        lightboxImg.src = img.src;
        caption.textContent = img.alt || '';
        
        lightbox.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        updateNavButtons();
    }
    
    function closeLightbox() {
        lightbox.classList.remove('active');
        document.body.style.overflow = '';
        
        // Reset image transform for next open
        setTimeout(() => {
            lightboxImg.src = '';
        }, 300);
    }
    
    function showPrev() {
        if (currentIndex > 0) {
            currentIndex--;
            updateImage();
        }
    }
    
    function showNext() {
        if (currentIndex < images.length - 1) {
            currentIndex++;
            updateImage();
        }
    }
    
    function updateImage() {
        // Animate out
        lightboxImg.style.transform = 'scale(0.8)';
        lightboxImg.style.opacity = '0';
        
        setTimeout(() => {
            const img = images[currentIndex];
            lightboxImg.src = img.src;
            caption.textContent = img.alt || '';
            
            // Animate in
            setTimeout(() => {
                lightboxImg.style.transform = 'scale(1)';
                lightboxImg.style.opacity = '1';
            }, 50);
            
            updateNavButtons();
        }, 200);
    }
    
    function updateNavButtons() {
        prevBtn.style.opacity = currentIndex === 0 ? '0.3' : '1';
        prevBtn.style.pointerEvents = currentIndex === 0 ? 'none' : 'auto';
        nextBtn.style.opacity = currentIndex === images.length - 1 ? '0.3' : '1';
        nextBtn.style.pointerEvents = currentIndex === images.length - 1 ? 'none' : 'auto';
    }
    
    // Event listeners
    closeBtn.addEventListener('click', closeLightbox);
    overlay.addEventListener('click', closeLightbox);
    prevBtn.addEventListener('click', showPrev);
    nextBtn.addEventListener('click', showNext);
    
    // Keyboard navigation
    document.addEventListener('keydown', (e) => {
        if (!lightbox.classList.contains('active')) return;
        
        switch (e.key) {
            case 'Escape':
                closeLightbox();
                break;
            case 'ArrowLeft':
                showPrev();
                break;
            case 'ArrowRight':
                showNext();
                break;
        }
    });
    
    // Add click handlers to story card images
    document.addEventListener('click', (e) => {
        const storyImage = e.target.closest('.story-card-image img');
        if (storyImage) {
            updateImages();
            const index = images.indexOf(storyImage);
            if (index !== -1) {
                openLightbox(index);
            }
        }
    });
}

// Activation Code Utilities
const ActivationCode = {
    // Generate a random activation code
    generate: function() {
        const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
        let code = '';
        for (let i = 0; i < 16; i++) {
            if (i > 0 && i % 4 === 0) code += '-';
            code += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return code;
    },
    
    // Validate activation code format
    validate: function(code) {
        const pattern = /^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/;
        return pattern.test(code.toUpperCase());
    },
    
    // Store activation code locally (for demo purposes)
    store: function(code) {
        const codes = this.getStoredCodes();
        if (!codes.includes(code)) {
            codes.push(code);
            localStorage.setItem('clearc_activation_codes', JSON.stringify(codes));
        }
    },
    
    // Get stored activation codes
    getStoredCodes: function() {
        const stored = localStorage.getItem('clearc_activation_codes');
        return stored ? JSON.parse(stored) : [];
    },
    
    // Check if code is valid (exists in storage)
    isValid: function(code) {
        const codes = this.getStoredCodes();
        return codes.includes(code.toUpperCase());
    }
};

// Payment handling - 通过 pay.php 处理，不在前端暴露任何支付配置
const Payment = {
    // 跳转到支付页面
    goToPay: function(method = 'alipay') {
        // 直接跳转到 pay.php，由服务端处理支付逻辑
        window.location.href = '/pay.php';
    },
    
    // 提交支付（通过表单）
    submitPay: function(method = 'alipay') {
        const form = document.createElement('form');
        form.method = 'POST';
        form.action = '/pay.php?action=pay';
        
        const input = document.createElement('input');
        input.type = 'hidden';
        input.name = 'method';
        input.value = method;
        
        form.appendChild(input);
        document.body.appendChild(form);
        form.submit();
    }
};

// Export for use in other scripts
window.ClearC = {
    ActivationCode,
    Payment
};
