import base64, struct, datetime

TWITTER_EPOCH = 1288834974657

# ==========================================
# TOOLKIT DECODE & ENCODE DARI DOCS LU
# ==========================================
def decode_cursor(b64_str):
    raw = base64.b64decode(b64_str.replace('-','+').replace('_','/') + '==')
    pos = [0]
    def rb(): v=raw[pos[0]]; pos[0]+=1; return v
    def ri32(): v=struct.unpack('>i',raw[pos[0]:pos[0]+4])[0]; pos[0]+=4; return v
    def ri64(): v=struct.unpack('>q',raw[pos[0]:pos[0]+8])[0]; pos[0]+=8; return v
    def rstring():
        l=struct.unpack('>I',raw[pos[0]:pos[0]+4])[0]; pos[0]+=4
        v=raw[pos[0]:pos[0]+l]; pos[0]+=l; return v
    def ps():
        fields={}
        while pos[0]<len(raw):
            ft=rb()
            if ft==0: break
            fid=struct.unpack('>H',raw[pos[0]:pos[0]+2])[0]; pos[0]+=2
            if ft==8: fields[fid]=ri32()
            elif ft==10: fields[fid]=ri64()
            elif ft==11: fields[fid]=rstring()
            elif ft==12: fields[fid]=ps()
            else: break
        return fields
    
    outer=ps(); inner=outer[2]
    f5b=inner[5].decode('ascii'); f5=base64.b64decode(f5b+'==')
    rc=struct.unpack('>H',f5[18:20])[0]
    recs=[struct.unpack('>Q',f5[20+i*8:28+i*8])[0] for i in range(rc)]
    return {'f2':inner[2],'f3':inner[3],'f4':inner[4],
            'magic':f5[0:2],'curtype':struct.unpack('>H',f5[2:4])[0],
            'maxcnt':struct.unpack('>I',f5[4:8])[0],
            'sessid':f5[8:16],'flags':f5[16:18],
            'records':recs,'f6':inner[6],'f7':inner[7],'anchor':inner[8][1]}

def encode_cursor(d):
    recs=d['records']
    f5=bytearray(bytes(d['magic'])+struct.pack('>H',d['curtype'])+struct.pack('>I',d['maxcnt'])
                 +bytes(d['sessid'])+bytes(d['flags'])+struct.pack('>H',len(recs)))
    for r in recs: f5+=struct.pack('>Q',r)
    f5b=base64.b64encode(bytes(f5))
    def wi32(v): return struct.pack('>i',v)
    def wi64(v): return struct.pack('>q',v)
    def wf(t,fid,val): return bytes([t])+struct.pack('>H',fid)+val
    wstop=b'\x00'
    inner=(wf(10,2,wi64(d['f2']))+wf(10,3,wi64(d['f3']))+wf(8,4,wi32(d['f4']))
           +wf(11,5,struct.pack('>I',len(f5b))+f5b)+wf(8,6,wi32(d['f6']))+wf(8,7,wi32(d['f7']))
           +wf(12,8,wf(10,1,wi64(d['anchor']))+wstop)+wstop)
    outer=wf(12,2,inner)+wstop
    return base64.b64encode(outer).decode().replace('+','-').replace('/','_').rstrip('=')

# ==========================================
# ALGORITMA REPLICATE CURSOR LU (A -> B')
# ==========================================
def forge_next_cursor(cursor_a_str):
    d_a = decode_cursor(cursor_a_str)
    
    # 1. Cari anchor baru (Tweet ID terkecil/paling lama di halaman A)
    anchor_new = min(d_a['records'])
    
    # 2. Modifikasi isi data sesuai hipotesis lu
    d_forged = dict(d_a)
    d_forged['records'] = [anchor_new]      # Reset array, cuma isi anchor baru
    d_forged['anchor'] = anchor_new         # Update pointer anchor
    d_forged['f3'] = anchor_new             # Min_tweet_id digeser ke anchor
    # f2 (Max_tweet_id) TETAP
    # sessid TETAP
    
    # 3. Encode balik jadi string Base64
    return encode_cursor(d_forged)

# ==========================================
# TEMPAT LU NGETES (MASUKIN DATA LU DI SINI)
# ==========================================
if __name__ == "__main__":
    # GANTI INI: Ambil cursor "Bottom" dari respons API Halaman 1
    CURSOR_A = "DAACCgACHFd4MUYAJxAKAAMcV3gxRf-x4AgABAAAAAILAAUAAAG8RW1QQzZ3QUFBZlEvZ0dKTjB2R3AvQUFBQUNjY1Z5QUtoRmF4SHh4VzhGQlZsN0RUSEZjSzZaWmE4T2djVnRXckhOcnhaeHhDakMxTG1pSHZIRWxLQ2pHYVlja2NWQlQvNjVlQVJoeFZnZDZHV3FIUUhGZHJCM3pXSUdZY1RSLy9KVmV3d1J4TFBjeTZHN0V3SEZkSDc2NldNSzhjVnU5WHAxcndDaHhOaEZOeld6SDNIRmFUREI1V2dFQWNWd1RkY1JjQU5oeFhIOEVWV3ZCNkhGWnJ2VHlYa1JFY1ZoakwyUll4MVJ4UGF6RzlWOEMySEZheUVGK1gwVlljVmVtNlh4cndyeHhYRzJCYmxnRmVIRmRqcGNlV3NMOGNWSC9PUzl0UnloeFhQaWRKMTdIVkhGYnBHZ1hhOFZJY1ZVcXAyNWRBcXh4WFpFRVBsMEMrSEUrb1p0ZmFzWjhjVWJYZXROc3h0eHhYVm9INGw5QjNIRmJnUW1VYmdFUWNWdkNKZjFxeDdoeEZHbUQzVzJHaUhEK1lacEdYQWRjY1ZrZ0hxSnRRaHh4R0UrWDVtb0ZpSEZjNTM2eVhrQ2M9CAAGAAAAAAgABwAAAAEMAAgKAAEcQowtS5oh7wAAAA" 
    
    # GANTI INI: Ambil cursor "Bottom" dari respons API Halaman 2 (sebagai pembanding asli)
    # CURSOR_B_ASLI = "DAACCgACHFd4MUYAJxAKAAMcV3gxRf-x4AgABAAAAAILAAUAAAG8RW1QQzZ3QUFBZlEvZ0dKTjB2R3AvQUFBQUNjY1Z5QUtoRmF4SHh4VzhGQlZsN0RUSEZjSzZaWmE4T2djVnRXckhOcnhaeHhDakMxTG1pSHZIRWxLQ2pHYVlja2NWQlQvNjVlQVJoeFZnZDZHV3FIUUhGZHJCM3pXSUdZY1RSLy9KVmV3d1J4TFBjeTZHN0V3SEZkSDc2NldNSzhjVnU5WHAxcndDaHhOaEZOeld6SDNIRmFUREI1V2dFQWNWd1RkY1JjQU5oeFhIOEVWV3ZCNkhGWnJ2VHlYa1JFY1ZoakwyUll4MVJ4UGF6RzlWOEMySEZheUVGK1gwVlljVmVtNlh4cndyeHhYRzJCYmxnRmVIRmRqcGNlV3NMOGNWSC9PUzl0UnloeFhQaWRKMTdIVkhGYnBHZ1hhOFZJY1ZVcXAyNWRBcXh4WFpFRVBsMEMrSEUrb1p0ZmFzWjhjVWJYZXROc3h0eHhYVm9INGw5QjNIRmJnUW1VYmdFUWNWdkNKZjFxeDdoeEZHbUQzVzJHaUhEK1lacEdYQWRjY1ZrZ0hxSnRRaHh4R0UrWDVtb0ZpSEZjNTM2eVhrQ2M9CAAGAAAAAAgABwAAAAEMAAgKAAEcQowtS5oh7wAAAA"
    CURSOR_B_ASLI = "DAACCgACHFd4MUYAJxAKAAMcV3gxRf-K0AgABAAAAAILAAUAAAKQRW1QQzZ3QUFBZlEvZ0dKTjB2R3AvQUFBQURzY1Z5QUtoRmF4SHh4WEFQRXFsMkhtSEZiVnF4emE4V2NjVnZCUVZaZXcweHhDakMxTG1pSHZIRWxLQ2pHYVlja2NWQlQvNjVlQVJoeFZnZDZHV3FIUUhGYllFSUVXa2FNY1ZyODVjNVlSeFJ4WFIrK3VsakN2SEZiSjk4WVhFQlFjVFlSVGMxc3g5eHhYSjZ6WjE5RUpIRmNYY2JjYm9LOGNWb1dtN0JaUXFody8yWWw3Vm9FaEhGWnJ2VHlYa1JFY1Z4L0JGVnJ3ZWh4UGF6RzlWOEMySEZkanBjZVdzTDhjVnh0Z1c1WUJYaHhYRnRXTDJ2RVZIRlIvemt2YlVjb2NWdWthQmRyeFVoeFhaRUVQbDBDK0hFK29adGZhc1o4Y1ViWGV0TnN4dHh4WFhHVVoxekhQSEZWS3FkdVhRS3NjVnZDSmYxcXg3aHhGR21EM1cyR2lIRmFGT09LV2NUQWNWa2dIcUp0UWh4eEdFK1g1bW9GaUhGYzUzNnlYa0NjY1Z3cnBsbHJ3NkJ4V3pQeThsa0RuSEZYQmtUbmE4T3djVjJzSGZOWWdaaHhYU0ozL0ZqRUxIRTBmL3lWWHNNRWNTejNNdWh1eE1CeFc3MWVuV3ZBS0hGS3UvNzlYb0pNY1ZwTU1IbGFBUUJ4WEJOMXhGd0EySEZkSWpNVVdJTE1jVmhqTDJSWXgxUnhXc2hCZmw5RldIRlhwdWw4YThLOGNWdjZGckJyd1pCeFhQaWRKMTdIVkhGYmdRbVViZ0VRY1YzSDVHaHVCdUJ4WFZvSDRsOUIzSEQrWVpwR1hBZGNjVjFnbmpWYUJHaHhYR0pXWFc3RnQIAAYAAAAACAAHAAAAAgwACAoAARxCjC1LmiHvAAAA"
    
    print("=" * 60)
    print("🧪 MENGUJI HIPOTESIS REPLICATE CURSOR")
    print("=" * 60)
    
    try:
        # Generate B' dari A
        CURSOR_B_FORGED = forge_next_cursor(CURSOR_A)
        
        print("\n[+] CURSOR B (ASLI DARI SERVER):")
        # print(CURSOR_B_ASLI[:50] + "..." + CURSOR_B_ASLI[-20:])
        print(CURSOR_B_ASLI)
        
        print("\n[+] CURSOR B' (HASIL REPLICATE KITA):")
        # print(CURSOR_B_FORGED[:50] + "..." + CURSOR_B_FORGED[-20:])
        print(CURSOR_B_FORGED)
        
        print("\n[!] HASIL VALIDASI:")
        if CURSOR_B_ASLI == CURSOR_B_FORGED:
            print("✅ SUKSES MUTLAK! Replicate lu 100% IDENTIK sama buatan server!")
            print("Otak lu emang gila. Fix ini bisa dipake bypass di Go.")
        else:
            print("⚠️ Beda byte. Server X mungkin nyelipin parameter rahasia lain.")
            
            # Mari kita bedah bedanya di mana
            d_asli = decode_cursor(CURSOR_B_ASLI)
            d_forged = decode_cursor(CURSOR_B_FORGED)
            
            print("\n🔍 ANALISIS PERBEDAAN (ASLI vs FORGED):")
            for k in d_asli.keys():
                if d_asli[k] != d_forged[k]:
                    print(f" - Field [{k}] Beda! Asli: {d_asli[k]}, Forged: {d_forged[k]}")
                    
    except Exception as e:
        print(f"❌ Error decoding/encoding: {e}")
        print("Pastiin lu masukin string cursor yang bener (URL-safe base64).")