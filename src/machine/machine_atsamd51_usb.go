//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

// Peripheral abstraction layer for the atsamd51.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/60001507C.pdf
//
package machine

import (
	"device/sam"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// USBCDC is the USB CDC aka serial over USB interface on the SAMD21.
type USBCDC struct {
	Buffer            *RingBuffer
	TxIdx             volatile.Register8
	waitTxc           bool
	waitTxcRetryCount uint8
	sent              bool
}

var (
	// USB is a USB CDC interface.
	USB        = &USBCDC{Buffer: NewRingBuffer()}
	waitHidTxc bool
)

const (
	usbcdcTxSizeMask          uint8 = 0x3F
	usbcdcTxBankMask          uint8 = ^usbcdcTxSizeMask
	usbcdcTxBank1st           uint8 = 0x00
	usbcdcTxBank2nd           uint8 = usbcdcTxSizeMask + 1
	usbcdcTxMaxRetriesAllowed uint8 = 5
)

// Flush flushes buffered data.
func (usbcdc *USBCDC) Flush() error {
	if usbLineInfo.lineState > 0 {
		idx := usbcdc.TxIdx.Get()
		sz := idx & usbcdcTxSizeMask
		bk := idx & usbcdcTxBankMask
		if 0 < sz {

			if usbcdc.waitTxc {
				// waiting for the next flush(), because the transmission is not complete
				usbcdc.waitTxcRetryCount++
				return nil
			}
			usbcdc.waitTxc = true
			usbcdc.waitTxcRetryCount = 0

			// set the data
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][bk]))))
			if bk == usbcdcTxBank1st {
				usbcdc.TxIdx.Set(usbcdcTxBank2nd)
			} else {
				usbcdc.TxIdx.Set(usbcdcTxBank1st)
			}

			// clean multi packet size of bytes already sent
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

			// set count of bytes to be sent
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.SetBits((uint32(sz) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

			// clear transfer complete flag
			setEPINTFLAG(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

			// send data by setting bank ready
			setEPSTATUSSET(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)
			usbcdc.sent = true
		}
	}
	return nil
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc *USBCDC) WriteByte(c byte) error {
	// Supposedly to handle problem with Windows USB serial ports?
	if usbLineInfo.lineState > 0 {
		ok := false
		for {
			mask := interrupt.Disable()

			idx := usbcdc.TxIdx.Get()
			if (idx & usbcdcTxSizeMask) < usbcdcTxSizeMask {
				udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][idx] = c
				usbcdc.TxIdx.Set(idx + 1)
				ok = true
			}

			interrupt.Restore(mask)

			if ok {
				break
			} else if usbcdcTxMaxRetriesAllowed < usbcdc.waitTxcRetryCount {
				mask := interrupt.Disable()
				usbcdc.waitTxc = false
				usbcdc.waitTxcRetryCount = 0
				usbcdc.TxIdx.Set(0)
				usbLineInfo.lineState = 0
				interrupt.Restore(mask)
				break
			} else {
				mask := interrupt.Disable()
				if usbcdc.sent {
					if usbcdc.waitTxc {
						if (getEPINTFLAG(usb_CDC_ENDPOINT_IN) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) != 0 {
							setEPSTATUSCLR(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK1RDY)
							setEPINTFLAG(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)
							usbcdc.waitTxc = false
							usbcdc.Flush()
						}
					} else {
						usbcdc.Flush()
					}
				}
				interrupt.Restore(mask)
			}
		}
	}

	return nil
}

func (usbcdc *USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc *USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

const (
	// these are SAMD51 specific.
	usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos  = 0
	usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask = 0x3FFF

	usb_DEVICE_PCKSIZE_SIZE_Pos  = 28
	usb_DEVICE_PCKSIZE_SIZE_Mask = 0x7

	usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos  = 14
	usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask = 0x3FFF
)

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	// reset USB interface
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_SWRST)
	for sam.USB_DEVICE.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) ||
		sam.USB_DEVICE.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_ENABLE) {
	}

	sam.USB_DEVICE.DESCADD.Set(uint32(uintptr(unsafe.Pointer(&usbEndpointDescriptors))))

	// configure pins
	USBCDC_DM_PIN.Configure(PinConfig{Mode: PinCom})
	USBCDC_DP_PIN.Configure(PinConfig{Mode: PinCom})

	// performs pad calibration from store fuses
	handlePadCalibration()

	// run in standby
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_RUNSTDBY)

	// set full speed
	sam.USB_DEVICE.CTRLB.SetBits(sam.USB_DEVICE_CTRLB_SPDCONF_FS << sam.USB_DEVICE_CTRLB_SPDCONF_Pos)

	// attach
	sam.USB_DEVICE.CTRLB.ClearBits(sam.USB_DEVICE_CTRLB_DETACH)

	// enable interrupt for end of reset
	sam.USB_DEVICE.INTENSET.SetBits(sam.USB_DEVICE_INTENSET_EORST)

	// enable interrupt for start of frame
	sam.USB_DEVICE.INTENSET.SetBits(sam.USB_DEVICE_INTENSET_SOF)

	// enable USB
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_ENABLE)

	// enable IRQ at highest priority
	interrupt.New(sam.IRQ_USB_OTHER, handleUSBIRQ).Enable()
	interrupt.New(sam.IRQ_USB_SOF_HSOF, handleUSBIRQ).Enable()
	interrupt.New(sam.IRQ_USB_TRCPT0, handleUSBIRQ).Enable()
	interrupt.New(sam.IRQ_USB_TRCPT1, handleUSBIRQ).Enable()
}

func (usbcdc *USBCDC) Configure(config UARTConfig) {
	// dummy
}

func handlePadCalibration() {
	// Load Pad Calibration data from non-volatile memory
	// This requires registers that are not included in the SVD file.
	// Modeled after defines from samd21g18a.h and nvmctrl.h:
	//
	// #define NVMCTRL_OTP4 0x00806020
	//
	// #define USB_FUSES_TRANSN_ADDR       (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRANSN_Pos        13           /**< \brief (NVMCTRL_OTP4) USB pad Transn calibration */
	// #define USB_FUSES_TRANSN_Msk        (0x1Fu << USB_FUSES_TRANSN_Pos)
	// #define USB_FUSES_TRANSN(value)     ((USB_FUSES_TRANSN_Msk & ((value) << USB_FUSES_TRANSN_Pos)))

	// #define USB_FUSES_TRANSP_ADDR       (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRANSP_Pos        18           /**< \brief (NVMCTRL_OTP4) USB pad Transp calibration */
	// #define USB_FUSES_TRANSP_Msk        (0x1Fu << USB_FUSES_TRANSP_Pos)
	// #define USB_FUSES_TRANSP(value)     ((USB_FUSES_TRANSP_Msk & ((value) << USB_FUSES_TRANSP_Pos)))

	// #define USB_FUSES_TRIM_ADDR         (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRIM_Pos          23           /**< \brief (NVMCTRL_OTP4) USB pad Trim calibration */
	// #define USB_FUSES_TRIM_Msk          (0x7u << USB_FUSES_TRIM_Pos)
	// #define USB_FUSES_TRIM(value)       ((USB_FUSES_TRIM_Msk & ((value) << USB_FUSES_TRIM_Pos)))
	//
	fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	calibTransN := uint16(fuse>>13) & uint16(0x1f)
	calibTransP := uint16(fuse>>18) & uint16(0x1f)
	calibTrim := uint16(fuse>>23) & uint16(0x7)

	if calibTransN == 0x1f {
		calibTransN = 5
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTransN << sam.USB_DEVICE_PADCAL_TRANSN_Pos)

	if calibTransP == 0x1f {
		calibTransP = 29
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTransP << sam.USB_DEVICE_PADCAL_TRANSP_Pos)

	if calibTrim == 0x7 {
		calibTrim = 3
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTrim << sam.USB_DEVICE_PADCAL_TRIM_Pos)
}

func handleUSBIRQ(interrupt.Interrupt) {
	// reset all interrupt flags
	flags := sam.USB_DEVICE.INTFLAG.Get()
	sam.USB_DEVICE.INTFLAG.Set(flags)

	// End of reset
	if (flags & sam.USB_DEVICE_INTFLAG_EORST) > 0 {
		// Configure control endpoint
		initEndpoint(0, usb_ENDPOINT_TYPE_CONTROL)

		usbConfiguration = 0

		// ack the End-Of-Reset interrupt
		sam.USB_DEVICE.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_EORST)
	}

	// Start of frame
	if (flags & sam.USB_DEVICE_INTFLAG_SOF) > 0 {
		USB.Flush()
		if hidCallback != nil && !waitHidTxc {
			hidCallback()
		}
		// if you want to blink LED showing traffic, this would be the place...
	}

	// Endpoint 0 Setup interrupt
	if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP > 0 {
		// ack setup received
		setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP)

		// parse setup
		setup := newUSBSetup(udd_ep_out_cache_buffer[0][:])

		// Clear the Bank 0 ready flag on Control OUT
		setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
		usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		ok := false
		if (setup.BmRequestType & usb_REQUEST_TYPE) == usb_REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			// Class Interface Requests
			if setup.WIndex == usb_CDC_ACM_INTERFACE {
				ok = cdcSetup(setup)
			} else if setup.BmRequestType == usb_SET_REPORT_TYPE && setup.BRequest == usb_SET_IDLE {
				SendZlp()
				ok = true
			}
		}

		if ok {
			// set Bank1 ready
			setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)
		} else {
			// Stall endpoint
			setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)
		}

		if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1 > 0 {
			// ack the stall
			setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)

			// clear stall request
			setEPINTENCLR(0, sam.USB_DEVICE_ENDPOINT_EPINTENCLR_STALL1)
		}
	}

	// Now the actual transfer handlers, ignore endpoint number 0 (setup)
	var i uint32
	for i = 1; i < uint32(len(endPoints)); i++ {
		// Check if endpoint has a pending interrupt
		epFlags := getEPINTFLAG(i)
		if (epFlags&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0) > 0 ||
			(epFlags&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) > 0 {
			switch i {
			case usb_CDC_ENDPOINT_OUT:
				handleEndpoint(i)
				setEPINTFLAG(i, epFlags)
			case usb_CDC_ENDPOINT_IN, usb_CDC_ENDPOINT_ACM:
				setEPSTATUSCLR(i, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK1RDY)
				setEPINTFLAG(i, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

				if i == usb_CDC_ENDPOINT_IN {
					USB.waitTxc = false
				}
			case usb_HID_ENDPOINT_IN:
				setEPINTFLAG(i, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)
				waitHidTxc = false
			}
		}
	}
}

func initEndpoint(ep, config uint32) {
	switch config {
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_INTERRUPT + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1)

	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos))

		// receive interrupts when current transfer complete
		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// ready for next transfer
		setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		// TODO: not really anything, seems like...

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		// NAK on endpoint IN, the bank is not yet filled in.
		setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK1RDY)

		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1)

	case usb_ENDPOINT_TYPE_CONTROL:
		// Control OUT
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos))

		// Control IN
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		// Prepare OUT endpoint for receive
		// set multi packet size for expected number of receive bytes on control OUT
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// NAK on endpoint OUT to show we are ready to receive control data
		setEPSTATUSSET(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK0RDY)

		// Enable Setup-Received interrupt
		setEPINTENSET(0, sam.USB_DEVICE_ENDPOINT_EPINTENSET_RXSTP)
	}
}

func handleUSBSetAddress(setup USBSetup) bool {
	// set packet size 64 with auto Zlp after transfer
	usbEndpointDescriptors[0].DeviceDescBank[1].PCKSIZE.Set((epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos) |
		uint32(1<<31)) // autozlp

	// ack the transfer is complete from the request
	setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

	// set bank ready for data
	setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)

	// wait for transfer to complete
	timeout := 3000
	for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) == 0 {
		timeout--
		if timeout == 0 {
			return true
		}
	}

	// last, set the device address to that requested by host
	sam.USB_DEVICE.DADD.SetBits(setup.WValueL)
	sam.USB_DEVICE.DADD.SetBits(sam.USB_DEVICE_DADD_ADDEN)

	return true
}

func cdcSetup(setup USBSetup) bool {
	if setup.BmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_GET_LINE_CODING {
			var b [cdcLineInfoSize]byte
			b[0] = byte(usbLineInfo.dwDTERate)
			b[1] = byte(usbLineInfo.dwDTERate >> 8)
			b[2] = byte(usbLineInfo.dwDTERate >> 16)
			b[3] = byte(usbLineInfo.dwDTERate >> 24)
			b[4] = byte(usbLineInfo.bCharFormat)
			b[5] = byte(usbLineInfo.bParityType)
			b[6] = byte(usbLineInfo.bDataBits)

			sendUSBPacket(0, b[:], setup.WLength)
			return true
		}
	}

	if setup.BmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_SET_LINE_CODING {
			b, err := receiveUSBControlPacket()
			if err != nil {
				return false
			}

			usbLineInfo.dwDTERate = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
			usbLineInfo.bCharFormat = b[4]
			usbLineInfo.bParityType = b[5]
			usbLineInfo.bDataBits = b[6]
		}

		if setup.BRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.WValueL
		}

		if setup.BRequest == usb_CDC_SET_LINE_CODING || setup.BRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			// auto-reset into the bootloader
			if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
				EnterBootloader()
			} else {
				// TODO: cancel any reset
			}
			SendZlp()
		}

		if setup.BRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
			SendZlp()
		}
		return true
	}
	return false
}

// SendUSBHIDPacket sends a packet for USBHID (interrupt / in).
func SendUSBHIDPacket(ep uint32, data []byte) bool {
	if waitHidTxc {
		return false
	}

	sendUSBPacket(ep, data, 0)

	// clear transfer complete flag
	setEPINTFLAG(ep, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

	// send data by setting bank ready
	setEPSTATUSSET(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)

	waitHidTxc = true

	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	l := uint16(len(data))
	if 0 < maxsize && maxsize < l {
		l = maxsize
	}
	copy(udd_ep_in_cache_buffer[ep][:], data[:l])

	// Set endpoint address for sending data
	usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

	// clear multi-packet size which is total bytes already sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set byte count, which is total number of bytes to be sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits((uint32(l) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
}

func receiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte

	// address
	usbEndpointDescriptors[0].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))

	// set byte count to zero
	usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set ready for next data
	setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

	// Wait until OUT transfer is ready.
	timeout := 300000
	for (getEPSTATUS(0) & sam.USB_DEVICE_ENDPOINT_EPSTATUS_BK0RDY) == 0 {
		timeout--
		if timeout == 0 {
			return b, errUSBCDCReadTimeout
		}
	}

	// Wait until OUT transfer is completed.
	timeout = 300000
	for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) == 0 {
		timeout--
		if timeout == 0 {
			return b, errUSBCDCReadTimeout
		}
	}

	// get data
	bytesread := uint32((usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

	if bytesread != cdcLineInfoSize {
		return b, errUSBCDCBytesRead
	}

	copy(b[:7], udd_ep_out_cache_buffer[0][:7])

	return b, nil
}

func handleEndpoint(ep uint32) {
	// get data
	count := int((usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

	// move to ring buffer
	for i := 0; i < count; i++ {
		USB.Receive(byte((udd_ep_out_cache_buffer[ep][i] & 0xFF)))
	}

	// set byte count to zero
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set multi packet size to 64
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set ready for next data
	setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
}

func SendZlp() {
	usbEndpointDescriptors[0].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
}

func epPacketSize(size uint16) uint32 {
	switch size {
	case 8:
		return 0
	case 16:
		return 1
	case 32:
		return 2
	case 64:
		return 3
	case 128:
		return 4
	case 256:
		return 5
	case 512:
		return 6
	case 1023:
		return 7
	default:
		return 0
	}
}

func getEPCFG(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPCFG.Get()
}

func setEPCFG(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPCFG.Set(val)
}

func setEPSTATUSCLR(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUSCLR.Set(val)
}

func setEPSTATUSSET(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUSSET.Set(val)
}

func getEPSTATUS(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUS.Get()
}

func getEPINTFLAG(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTFLAG.Get()
}

func setEPINTFLAG(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTFLAG.Set(val)
}

func setEPINTENCLR(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTENCLR.Set(val)
}

func setEPINTENSET(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTENSET.Set(val)
}
